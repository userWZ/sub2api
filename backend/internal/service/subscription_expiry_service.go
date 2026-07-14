package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/google/uuid"
)

const (
	// subscriptionExpiryReminderLeaderLockKey gates the per-cycle reminder scan so
	// that only one instance walks all active subscriptions and sends reminder
	// emails, avoiding redundant full scans and duplicate emails.
	subscriptionExpiryReminderLeaderLockKey = "subscription:expiry:reminder:leader"
	// subscriptionExpiryReminderLeaderLockTTL bounds crash recovery; the scan can
	// page through many subscriptions, so keep it comfortably above one cycle.
	subscriptionExpiryReminderLeaderLockTTL = 5 * time.Minute
)

// SubscriptionExpiryService periodically updates expired subscription status.
type SubscriptionExpiryService struct {
	userSubRepo              UserSubscriptionRepository
	settingRepo              SettingRepository
	notificationEmailService *NotificationEmailService
	interval                 time.Duration
	stopCh                   chan struct{}
	stopOnce                 sync.Once
	wg                       sync.WaitGroup

	lockCache  LeaderLockCache
	db         *sql.DB
	instanceID string
}

func NewSubscriptionExpiryService(userSubRepo UserSubscriptionRepository, interval time.Duration) *SubscriptionExpiryService {
	return &SubscriptionExpiryService{
		userSubRepo: userSubRepo,
		interval:    interval,
		stopCh:      make(chan struct{}),
		instanceID:  uuid.NewString(),
	}
}

// SetLeaderLock injects the leader-lock cache and DB used to elect a single
// instance for the periodic expiry-reminder scan. When both are nil the scan runs
// ungated (single-instance / test behavior).
func (s *SubscriptionExpiryService) SetLeaderLock(lockCache LeaderLockCache, db *sql.DB) {
	if s == nil {
		return
	}
	s.lockCache = lockCache
	s.db = db
}

func (s *SubscriptionExpiryService) SetSettingRepository(settingRepo SettingRepository) {
	s.settingRepo = settingRepo
}

func (s *SubscriptionExpiryService) SetNotificationEmailService(notificationEmailService *NotificationEmailService) {
	s.notificationEmailService = notificationEmailService
}

func (s *SubscriptionExpiryService) Start() {
	if s == nil || s.userSubRepo == nil || s.interval <= 0 {
		return
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		s.runOnce()
		for {
			select {
			case <-ticker.C:
				s.runOnce()
			case <-s.stopCh:
				return
			}
		}
	}()
}

func (s *SubscriptionExpiryService) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() {
		close(s.stopCh)
	})
	s.wg.Wait()
}

func (s *SubscriptionExpiryService) runOnce() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	updated, err := s.userSubRepo.BatchUpdateExpiredStatus(ctx)
	if err != nil {
		log.Printf("[SubscriptionExpiry] Update expired subscriptions failed: %v", err)
		return
	}
	if updated > 0 {
		log.Printf("[SubscriptionExpiry] Updated %d expired subscriptions", updated)
	}
	s.sendExpiryReminders(ctx)
}

func (s *SubscriptionExpiryService) sendExpiryReminders(ctx context.Context) {
	if s == nil || s.userSubRepo == nil || s.notificationEmailService == nil {
		return
	}
	if !s.expiryReminderEnabled(ctx) {
		return
	}

	// Multi-instance guard: only the leader walks every active subscription and
	// sends reminders, avoiding N× full scans and duplicate reminder emails.
	release, ok := tryAcquireSingletonLeaderLock(ctx, s.lockCache, s.db, subscriptionExpiryReminderLeaderLockKey, s.instanceID, subscriptionExpiryReminderLeaderLockTTL)
	if !ok {
		return
	}
	defer release()
	renewalOffer := s.loadRenewalEmailOffer(ctx)
	for page := 1; ; page++ {
		subs, pag, err := s.userSubRepo.List(ctx, pagination.PaginationParams{Page: page, PageSize: 200}, nil, nil, SubscriptionStatusActive, "", "expires_at", "asc")
		if err != nil {
			log.Printf("[SubscriptionExpiry] List active subscriptions for reminder failed: %v", err)
			return
		}
		for i := range subs {
			s.sendExpiryReminderIfDue(ctx, &subs[i])
			s.sendRenewalOfferIfDue(ctx, &subs[i], renewalOffer)
		}
		if pag == nil || page >= pag.Pages || len(subs) == 0 {
			return
		}
	}
}

type renewalEmailOffer struct {
	Enabled         bool
	WindowDays      int
	DiscountPercent float64
	RolloverPercent float64
}

func (s *SubscriptionExpiryService) loadRenewalEmailOffer(ctx context.Context) renewalEmailOffer {
	offer := renewalEmailOffer{WindowDays: 14, DiscountPercent: 10, RolloverPercent: 20}
	if s == nil || s.settingRepo == nil {
		return offer
	}
	values, err := s.settingRepo.GetMultiple(ctx, []string{
		SettingRenewalOfferEnabled, SettingRenewalEmailEnabled, SettingRenewalWindowDays,
		SettingRenewalDiscountPercent, SettingRenewalRolloverPercent,
	})
	if err != nil {
		return offer
	}
	offer.Enabled = values[SettingRenewalOfferEnabled] == "true" && values[SettingRenewalEmailEnabled] != "false"
	offer.WindowDays = pcParseInt(values[SettingRenewalWindowDays], 14)
	offer.DiscountPercent = pcParseFloat(values[SettingRenewalDiscountPercent], 10)
	offer.RolloverPercent = pcParseFloat(values[SettingRenewalRolloverPercent], 20)
	return offer
}

func (s *SubscriptionExpiryService) sendRenewalOfferIfDue(ctx context.Context, sub *UserSubscription, offer renewalEmailOffer) {
	daysRemaining := 0
	if sub != nil {
		daysRemaining = sub.DaysRemaining()
	}
	if !offer.Enabled || sub == nil || sub.User == nil || sub.Group == nil || sub.User.Email == "" || (daysRemaining != offer.WindowDays && daysRemaining != offer.WindowDays-1) {
		return
	}
	baseURL := strings.TrimRight(s.notificationEmailService.baseURL(ctx), "/")
	renewURL := baseURL + "/payment?tab=subscription"
	if baseURL == "" {
		renewURL = "/payment?tab=subscription"
	}
	if err := s.notificationEmailService.Send(ctx, NotificationEmailSendInput{
		Event:          NotificationEmailEventSubscriptionRenewalOffer,
		RecipientEmail: sub.User.Email,
		RecipientName:  firstNonEmpty(sub.User.Username, sub.User.Email),
		UserID:         sub.UserID,
		SourceType:     "user_subscription",
		SourceID:       strconv.FormatInt(sub.ID, 10),
		ReminderKey:    fmt.Sprintf("window-%dd", offer.WindowDays),
		Variables: map[string]string{
			"subscription_group": sub.Group.Name,
			"expiry_time":        sub.ExpiresAt.Format("2006-01-02 15:04"),
			"window_days":        strconv.Itoa(offer.WindowDays),
			"discount_percent":   strconv.FormatFloat(offer.DiscountPercent, 'f', -1, 64),
			"rollover_percent":   strconv.FormatFloat(offer.RolloverPercent, 'f', -1, 64),
			"renew_url":          renewURL,
		},
	}); err != nil {
		log.Printf("[SubscriptionExpiry] Send renewal offer failed: subscription=%d user=%d err=%v", sub.ID, sub.UserID, err)
	}
}

func (s *SubscriptionExpiryService) expiryReminderEnabled(ctx context.Context) bool {
	if s == nil || s.settingRepo == nil {
		return true
	}
	value, err := s.settingRepo.GetValue(ctx, SettingKeySubscriptionExpiryNotifyEnabled)
	if err != nil {
		if errors.Is(err, ErrSettingNotFound) {
			return true
		}
		log.Printf("[SubscriptionExpiry] Read expiry reminder switch failed: %v", err)
		return false
	}
	return !isFalseSettingValue(value)
}

func (s *SubscriptionExpiryService) sendExpiryReminderIfDue(ctx context.Context, sub *UserSubscription) {
	if sub == nil || sub.User == nil || sub.Group == nil || sub.User.Email == "" {
		return
	}
	daysRemaining := sub.DaysRemaining()
	if daysRemaining != 7 && daysRemaining != 3 && daysRemaining != 1 {
		return
	}
	if err := s.notificationEmailService.Send(ctx, NotificationEmailSendInput{
		Event:          NotificationEmailEventSubscriptionExpiryReminder,
		RecipientEmail: sub.User.Email,
		RecipientName:  firstNonEmpty(sub.User.Username, sub.User.Email),
		UserID:         sub.UserID,
		SourceType:     "user_subscription",
		SourceID:       strconv.FormatInt(sub.ID, 10),
		ReminderKey:    fmt.Sprintf("%dd", daysRemaining),
		Variables: map[string]string{
			"subscription_group": sub.Group.Name,
			"expiry_time":        sub.ExpiresAt.Format("2006-01-02 15:04"),
			"days_remaining":     strconv.Itoa(daysRemaining),
		},
	}); err != nil {
		log.Printf("[SubscriptionExpiry] Send expiry reminder failed: subscription=%d user=%d err=%v", sub.ID, sub.UserID, err)
	}
}
