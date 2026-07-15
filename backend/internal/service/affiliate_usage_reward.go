package service

import (
	"context"
	"math"
	"strconv"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	AffiliateUsageRewardAmountDefault       = 20.0
	AffiliateRewardInviterDailyLimitDefault = 5
	AffiliateRewardInviter30DayLimitDefault = 30
	AffiliateRewardIPDailyLimitDefault      = 2
	AffiliateRewardLimitMax                 = 100000
	AffiliateRewardReviewRemarkMaxLength    = 500
)

var (
	ErrAffiliateUsageRewardNotFound = infraerrors.NotFound("AFFILIATE_USAGE_REWARD_NOT_FOUND", "affiliate usage reward not found")
	ErrAffiliateUsageRewardReviewed = infraerrors.Conflict("AFFILIATE_USAGE_REWARD_REVIEWED", "affiliate usage reward has already been reviewed")
)

// AffiliateSettings is owned by the dedicated affiliate settings page. Amount
// values are system billing credits; they are not cash or a refund amount.
type AffiliateSettings struct {
	AffiliateEnabled               bool    `json:"affiliate_enabled"`
	RebateRate                     float64 `json:"rebate_rate"`
	RebateFreezeHours              int     `json:"rebate_freeze_hours"`
	RebateDurationDays             int     `json:"rebate_duration_days"`
	RebatePerInviteeCap            float64 `json:"rebate_per_invitee_cap"`
	DiscountEnabled                bool    `json:"discount_enabled"`
	DiscountMaxPercent             float64 `json:"discount_max_percent"`
	DiscountMinPayAmount           float64 `json:"discount_min_pay_amount"`
	UsageRewardEnabled             bool    `json:"usage_reward_enabled"`
	UsageRewardAmount              float64 `json:"usage_reward_amount"`
	RewardInviterDailyLimit        int     `json:"reward_inviter_daily_limit"`
	RewardInviterRolling30DayLimit int     `json:"reward_inviter_30d_limit"`
	RewardIPDailyLimit             int     `json:"reward_ip_daily_limit"`
	RewardReviewSameIP             bool    `json:"reward_review_same_ip"`
	RewardReviewMissingIP          bool    `json:"reward_review_missing_ip"`
}

type AffiliateUsageRewardEvent struct {
	InviteeUserID int64
	RequestID     string
	APIKeyID      int64
	IPAddress     string
	ActualCost    float64
}

type AffiliateUsageRewardProcessResult struct {
	Created    bool   `json:"created"`
	RewardID   int64  `json:"reward_id,omitempty"`
	InviterID  int64  `json:"inviter_id,omitempty"`
	Status     string `json:"status,omitempty"`
	RiskReason string `json:"risk_reason,omitempty"`
}

type AffiliateUsageRewardFilter struct {
	Search   string
	Status   string
	Page     int
	PageSize int
	StartAt  *time.Time
	EndAt    *time.Time
}

type AffiliateUsageRewardRecord struct {
	ID                int64      `json:"id"`
	InviterID         int64      `json:"inviter_id"`
	InviterEmail      string     `json:"inviter_email"`
	InviterUsername   string     `json:"inviter_username"`
	InviteeID         int64      `json:"invitee_id"`
	InviteeEmail      string     `json:"invitee_email"`
	InviteeUsername   string     `json:"invitee_username"`
	Amount            float64    `json:"amount"`
	Status            string     `json:"status"`
	RiskReason        string     `json:"risk_reason"`
	TriggerRequestID  string     `json:"trigger_request_id"`
	TriggerIPAddress  *string    `json:"trigger_ip,omitempty"`
	TriggerActualCost float64    `json:"trigger_actual_cost"`
	ReviewerUserID    *int64     `json:"reviewer_user_id,omitempty"`
	ReviewerEmail     *string    `json:"reviewer_email,omitempty"`
	ReviewRemark      string     `json:"review_remark"`
	ReviewedAt        *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}

type AffiliateUsageRewardReviewResult struct {
	RewardID  int64   `json:"reward_id"`
	InviterID int64   `json:"inviter_id"`
	Status    string  `json:"status"`
	Amount    float64 `json:"amount"`
}

// AffiliateUsageRewardRepository is an optional extension to the established
// affiliate repository contract, keeping existing integrations source-compatible.
type AffiliateUsageRewardRepository interface {
	ProcessUsageReward(ctx context.Context, event AffiliateUsageRewardEvent, settings AffiliateSettings) (*AffiliateUsageRewardProcessResult, error)
	ListUsageRewardRecords(ctx context.Context, filter AffiliateUsageRewardFilter) ([]AffiliateUsageRewardRecord, int64, error)
	ReviewUsageReward(ctx context.Context, rewardID, reviewerUserID int64, approve bool, remark string) (*AffiliateUsageRewardReviewResult, error)
}

func (s *AffiliateService) GetSettings(ctx context.Context) (*AffiliateSettings, error) {
	if s == nil || s.settingService == nil || s.settingService.settingRepo == nil {
		return nil, infraerrors.ServiceUnavailable("SERVICE_UNAVAILABLE", "affiliate settings unavailable")
	}
	keys := []string{
		SettingKeyAffiliateEnabled,
		SettingKeyAffiliateRebateRate,
		SettingKeyAffiliateRebateFreezeHours,
		SettingKeyAffiliateRebateDurationDays,
		SettingKeyAffiliateRebatePerInviteeCap,
		SettingKeyAffiliateDiscountEnabled,
		SettingKeyAffiliateDiscountMaxPercent,
		SettingKeyAffiliateDiscountMinPayAmount,
		SettingKeyAffiliateUsageRewardEnabled,
		SettingKeyAffiliateUsageRewardAmount,
		SettingKeyAffiliateRewardInviterDailyLimit,
		SettingKeyAffiliateRewardInviter30DayLimit,
		SettingKeyAffiliateRewardIPDailyLimit,
		SettingKeyAffiliateRewardReviewSameIP,
		SettingKeyAffiliateRewardReviewMissingIP,
	}
	values, err := s.settingService.settingRepo.GetMultiple(ctx, keys)
	if err != nil {
		return nil, err
	}
	settings := &AffiliateSettings{
		AffiliateEnabled:               parseBoolDefault(values[SettingKeyAffiliateEnabled], AffiliateEnabledDefault),
		RebateRate:                     parseFloatDefault(values[SettingKeyAffiliateRebateRate], AffiliateRebateRateDefault),
		RebateFreezeHours:              parseIntDefault(values[SettingKeyAffiliateRebateFreezeHours], AffiliateRebateFreezeHoursDefault),
		RebateDurationDays:             parseIntDefault(values[SettingKeyAffiliateRebateDurationDays], AffiliateRebateDurationDaysDefault),
		RebatePerInviteeCap:            parseFloatDefault(values[SettingKeyAffiliateRebatePerInviteeCap], AffiliateRebatePerInviteeCapDefault),
		DiscountEnabled:                parseBoolDefault(values[SettingKeyAffiliateDiscountEnabled], AffiliateDiscountEnabledDefault),
		DiscountMaxPercent:             parseFloatDefault(values[SettingKeyAffiliateDiscountMaxPercent], AffiliateDiscountMaxPercentDefault),
		DiscountMinPayAmount:           parseFloatDefault(values[SettingKeyAffiliateDiscountMinPayAmount], AffiliateDiscountMinPayAmountDefault),
		UsageRewardEnabled:             parseBoolDefault(values[SettingKeyAffiliateUsageRewardEnabled], false),
		UsageRewardAmount:              parseFloatDefault(values[SettingKeyAffiliateUsageRewardAmount], AffiliateUsageRewardAmountDefault),
		RewardInviterDailyLimit:        parseIntDefault(values[SettingKeyAffiliateRewardInviterDailyLimit], AffiliateRewardInviterDailyLimitDefault),
		RewardInviterRolling30DayLimit: parseIntDefault(values[SettingKeyAffiliateRewardInviter30DayLimit], AffiliateRewardInviter30DayLimitDefault),
		RewardIPDailyLimit:             parseIntDefault(values[SettingKeyAffiliateRewardIPDailyLimit], AffiliateRewardIPDailyLimitDefault),
		RewardReviewSameIP:             parseBoolDefault(values[SettingKeyAffiliateRewardReviewSameIP], true),
		RewardReviewMissingIP:          parseBoolDefault(values[SettingKeyAffiliateRewardReviewMissingIP], true),
	}
	settings.normalize()
	return settings, nil
}

func (s *AffiliateService) UpdateSettings(ctx context.Context, settings AffiliateSettings) (*AffiliateSettings, error) {
	if s == nil || s.settingService == nil || s.settingService.settingRepo == nil {
		return nil, infraerrors.ServiceUnavailable("SERVICE_UNAVAILABLE", "affiliate settings unavailable")
	}
	if err := settings.validate(); err != nil {
		return nil, err
	}
	settings.normalize()
	updates := map[string]string{
		SettingKeyAffiliateEnabled:                 strconv.FormatBool(settings.AffiliateEnabled),
		SettingKeyAffiliateRebateRate:              strconv.FormatFloat(settings.RebateRate, 'f', 8, 64),
		SettingKeyAffiliateRebateFreezeHours:       strconv.Itoa(settings.RebateFreezeHours),
		SettingKeyAffiliateRebateDurationDays:      strconv.Itoa(settings.RebateDurationDays),
		SettingKeyAffiliateRebatePerInviteeCap:     strconv.FormatFloat(settings.RebatePerInviteeCap, 'f', 8, 64),
		SettingKeyAffiliateDiscountEnabled:         strconv.FormatBool(settings.DiscountEnabled),
		SettingKeyAffiliateDiscountMaxPercent:      strconv.FormatFloat(settings.DiscountMaxPercent, 'f', 8, 64),
		SettingKeyAffiliateDiscountMinPayAmount:    strconv.FormatFloat(settings.DiscountMinPayAmount, 'f', 8, 64),
		SettingKeyAffiliateUsageRewardEnabled:      strconv.FormatBool(settings.UsageRewardEnabled),
		SettingKeyAffiliateUsageRewardAmount:       strconv.FormatFloat(settings.UsageRewardAmount, 'f', 8, 64),
		SettingKeyAffiliateRewardInviterDailyLimit: strconv.Itoa(settings.RewardInviterDailyLimit),
		SettingKeyAffiliateRewardInviter30DayLimit: strconv.Itoa(settings.RewardInviterRolling30DayLimit),
		SettingKeyAffiliateRewardIPDailyLimit:      strconv.Itoa(settings.RewardIPDailyLimit),
		SettingKeyAffiliateRewardReviewSameIP:      strconv.FormatBool(settings.RewardReviewSameIP),
		SettingKeyAffiliateRewardReviewMissingIP:   strconv.FormatBool(settings.RewardReviewMissingIP),
	}
	if err := s.settingService.settingRepo.SetMultiple(ctx, updates); err != nil {
		return nil, err
	}
	if s.settingService.onUpdate != nil {
		s.settingService.onUpdate()
	}
	return &settings, nil
}

func (s *AffiliateService) ProcessFirstUsageReward(ctx context.Context, event AffiliateUsageRewardEvent) (*AffiliateUsageRewardProcessResult, error) {
	if s == nil || s.repo == nil || event.InviteeUserID <= 0 || event.ActualCost <= 0 || math.IsNaN(event.ActualCost) || math.IsInf(event.ActualCost, 0) {
		return nil, nil
	}
	settings, err := s.GetSettings(ctx)
	if err != nil {
		return nil, err
	}
	if !settings.AffiliateEnabled || !settings.UsageRewardEnabled {
		return nil, nil
	}
	event.RequestID = strings.TrimSpace(event.RequestID)
	event.IPAddress = strings.TrimSpace(event.IPAddress)
	rewardRepo, ok := s.repo.(AffiliateUsageRewardRepository)
	if !ok {
		return nil, infraerrors.ServiceUnavailable("SERVICE_UNAVAILABLE", "affiliate usage reward repository unavailable")
	}
	result, err := rewardRepo.ProcessUsageReward(ctx, event, *settings)
	if err != nil {
		return nil, err
	}
	if result != nil && result.Created && result.Status == "granted" {
		s.invalidateAffiliateCaches(ctx, result.InviterID)
	}
	return result, nil
}

func (s *AffiliateService) AdminListUsageRewards(ctx context.Context, filter AffiliateUsageRewardFilter) ([]AffiliateUsageRewardRecord, int64, error) {
	if s == nil || s.repo == nil {
		return nil, 0, infraerrors.ServiceUnavailable("SERVICE_UNAVAILABLE", "affiliate service unavailable")
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}
	filter.Search = strings.TrimSpace(filter.Search)
	switch strings.TrimSpace(filter.Status) {
	case "pending", "granted", "rejected":
		filter.Status = strings.TrimSpace(filter.Status)
	default:
		filter.Status = ""
	}
	rewardRepo, ok := s.repo.(AffiliateUsageRewardRepository)
	if !ok {
		return nil, 0, infraerrors.ServiceUnavailable("SERVICE_UNAVAILABLE", "affiliate usage reward repository unavailable")
	}
	return rewardRepo.ListUsageRewardRecords(ctx, filter)
}

func (s *AffiliateService) AdminReviewUsageReward(ctx context.Context, rewardID, reviewerUserID int64, approve bool, remark string) (*AffiliateUsageRewardReviewResult, error) {
	if rewardID <= 0 || reviewerUserID <= 0 {
		return nil, infraerrors.BadRequest("INVALID_REWARD_REVIEW", "invalid reward review request")
	}
	remark = strings.TrimSpace(remark)
	if len(remark) > AffiliateRewardReviewRemarkMaxLength {
		return nil, infraerrors.BadRequest("INVALID_REMARK", "remark is too long")
	}
	rewardRepo, ok := s.repo.(AffiliateUsageRewardRepository)
	if !ok {
		return nil, infraerrors.ServiceUnavailable("SERVICE_UNAVAILABLE", "affiliate usage reward repository unavailable")
	}
	result, err := rewardRepo.ReviewUsageReward(ctx, rewardID, reviewerUserID, approve, remark)
	if err != nil {
		return nil, err
	}
	if result != nil && result.Status == "granted" {
		s.invalidateAffiliateCaches(ctx, result.InviterID)
	}
	return result, nil
}

func (s *AffiliateSettings) validate() error {
	if s == nil {
		return infraerrors.BadRequest("INVALID_AFFILIATE_SETTINGS", "affiliate settings are required")
	}
	finiteNonNegative := func(v float64) bool { return v >= 0 && !math.IsNaN(v) && !math.IsInf(v, 0) }
	if !finiteNonNegative(s.RebateRate) || s.RebateRate > AffiliateRebateRateMax ||
		!finiteNonNegative(s.RebatePerInviteeCap) ||
		!finiteNonNegative(s.DiscountMaxPercent) || s.DiscountMaxPercent > AffiliateDiscountMaxPercentMax ||
		!finiteNonNegative(s.DiscountMinPayAmount) ||
		!finiteNonNegative(s.UsageRewardAmount) || s.UsageRewardAmount <= 0 {
		return infraerrors.BadRequest("INVALID_AFFILIATE_SETTINGS", "affiliate amount or percentage is out of range")
	}
	if s.RebateFreezeHours < 0 || s.RebateFreezeHours > AffiliateRebateFreezeHoursMax ||
		s.RebateDurationDays < 0 || s.RebateDurationDays > AffiliateRebateDurationDaysMax ||
		s.RewardInviterDailyLimit < 0 || s.RewardInviterDailyLimit > AffiliateRewardLimitMax ||
		s.RewardInviterRolling30DayLimit < 0 || s.RewardInviterRolling30DayLimit > AffiliateRewardLimitMax ||
		s.RewardIPDailyLimit < 0 || s.RewardIPDailyLimit > AffiliateRewardLimitMax {
		return infraerrors.BadRequest("INVALID_AFFILIATE_SETTINGS", "affiliate limit is out of range")
	}
	return nil
}

func (s *AffiliateSettings) normalize() {
	if s == nil {
		return
	}
	s.RebateRate = clampAffiliateRebateRate(s.RebateRate)
	s.DiscountMaxPercent = clampAffiliateDiscountMaxPercent(s.DiscountMaxPercent)
	if s.RebateFreezeHours < 0 {
		s.RebateFreezeHours = 0
	}
	if s.RebateDurationDays < 0 {
		s.RebateDurationDays = 0
	}
	if s.RebatePerInviteeCap < 0 {
		s.RebatePerInviteeCap = 0
	}
	if s.DiscountMinPayAmount < 0 {
		s.DiscountMinPayAmount = 0
	}
	if s.UsageRewardAmount <= 0 || math.IsNaN(s.UsageRewardAmount) || math.IsInf(s.UsageRewardAmount, 0) {
		s.UsageRewardAmount = AffiliateUsageRewardAmountDefault
	}
	s.UsageRewardAmount = roundTo(s.UsageRewardAmount, 8)
	s.RewardInviterDailyLimit = clampNonNegativeInt(s.RewardInviterDailyLimit)
	s.RewardInviterRolling30DayLimit = clampNonNegativeInt(s.RewardInviterRolling30DayLimit)
	s.RewardIPDailyLimit = clampNonNegativeInt(s.RewardIPDailyLimit)
}

func parseBoolDefault(raw string, fallback bool) bool {
	v, err := strconv.ParseBool(strings.TrimSpace(raw))
	if err != nil {
		return fallback
	}
	return v
}

func parseFloatDefault(raw string, fallback float64) float64 {
	v, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
		return fallback
	}
	return v
}

func parseIntDefault(raw string, fallback int) int {
	v, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return fallback
	}
	return v
}

func clampNonNegativeInt(v int) int {
	if v < 0 {
		return 0
	}
	if v > AffiliateRewardLimitMax {
		return AffiliateRewardLimitMax
	}
	return v
}
