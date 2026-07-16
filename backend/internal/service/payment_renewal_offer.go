package service

import (
	"context"
	"math"
	"time"
)

type renewalOfferQuote struct {
	Eligible             bool
	DiscountPercent      float64
	DiscountAmount       float64
	RolloverAmount       float64
	SourceSubscriptionID *int64
	SourceExpiresAt      *time.Time
}

type RenewalOfferCheckoutSettings struct {
	Enabled          bool    `json:"enabled"`
	WindowDays       int     `json:"window_days"`
	BeforeExpiryDays int     `json:"before_expiry_days"`
	AfterExpiryDays  int     `json:"after_expiry_days"`
	DiscountEnabled  bool    `json:"discount_enabled"`
	DiscountPercent  float64 `json:"discount_percent"`
	RolloverEnabled  bool    `json:"rollover_enabled"`
	RolloverPercent  float64 `json:"rollover_percent"`
}

type RenewalOfferQuote struct {
	Eligible         bool    `json:"eligible"`
	DiscountPercent  float64 `json:"discount_percent"`
	DiscountedAmount float64 `json:"discounted_amount"`
	RolloverAmount   float64 `json:"rollover_amount"`
}

func (s *PaymentService) QuoteRenewalOffer(ctx context.Context, userID, groupID int64, planPrice float64, cfg *PaymentConfig) RenewalOfferQuote {
	quote := s.quoteRenewalOffer(ctx, userID, groupID, planPrice, cfg, time.Now())
	return RenewalOfferQuote{
		Eligible:         quote.Eligible,
		DiscountPercent:  quote.DiscountPercent,
		DiscountedAmount: roundRenewalCurrency(math.Max(planPrice-quote.DiscountAmount, 0)),
		RolloverAmount:   quote.RolloverAmount,
	}
}

func (s *PaymentService) RenewalOfferCheckoutSettings(ctx context.Context) RenewalOfferCheckoutSettings {
	settings := RenewalOfferCheckoutSettings{}
	if s == nil || s.configService == nil {
		return settings
	}
	cfg, err := s.configService.GetPaymentConfig(ctx)
	if err != nil {
		return settings
	}
	settings.Enabled = cfg.RenewalOfferEnabled
	settings.WindowDays = max(cfg.RenewalBeforeDays, cfg.RenewalAfterDays)
	settings.BeforeExpiryDays = cfg.RenewalBeforeDays
	settings.AfterExpiryDays = cfg.RenewalAfterDays
	settings.DiscountEnabled = cfg.RenewalDiscountEnabled
	settings.DiscountPercent = cfg.RenewalDiscountPercent
	settings.RolloverEnabled = cfg.RenewalRolloverEnabled
	settings.RolloverPercent = cfg.RenewalRolloverPercent
	return settings
}

func (s *PaymentService) quoteRenewalOffer(ctx context.Context, userID, groupID int64, planPrice float64, cfg *PaymentConfig, now time.Time) renewalOfferQuote {
	quote := renewalOfferQuote{}
	if s == nil || cfg == nil || !cfg.RenewalOfferEnabled || s.subscriptionSvc == nil || s.subscriptionSvc.userSubRepo == nil {
		return quote
	}
	sub, err := s.subscriptionSvc.userSubRepo.GetByUserIDAndGroupID(ctx, userID, groupID)
	if err != nil || sub == nil {
		return quote
	}
	beforeWindow := time.Duration(max(cfg.RenewalBeforeDays, 0)) * 24 * time.Hour
	afterWindow := time.Duration(max(cfg.RenewalAfterDays, 0)) * 24 * time.Hour
	if now.Before(sub.ExpiresAt.Add(-beforeWindow)) || now.After(sub.ExpiresAt.Add(afterWindow)) {
		return quote
	}

	if cfg.RenewalDiscountEnabled && planPrice >= cfg.RenewalDiscountMinOrder {
		quote.DiscountPercent = clampPercent(cfg.RenewalDiscountPercent)
		quote.DiscountAmount = roundRenewalCurrency(planPrice * quote.DiscountPercent / 100)
		if cfg.RenewalDiscountMaxAmount > 0 {
			quote.DiscountAmount = math.Min(quote.DiscountAmount, cfg.RenewalDiscountMaxAmount)
		}
	}
	if cfg.RenewalRolloverEnabled && sub.Group != nil && sub.Group.MonthlyLimitUSD != nil && *sub.Group.MonthlyLimitUSD > 0 {
		remaining := math.Max(*sub.Group.MonthlyLimitUSD-sub.MonthlyUsageUSD, 0)
		if remaining >= cfg.RenewalRolloverMinUnused {
			quote.RolloverAmount = roundRenewalAmount(remaining * clampPercent(cfg.RenewalRolloverPercent) / 100)
			if cfg.RenewalRolloverMaxAmount > 0 {
				quote.RolloverAmount = math.Min(quote.RolloverAmount, cfg.RenewalRolloverMaxAmount)
			}
		}
	}
	quote.Eligible = quote.DiscountAmount > 0 || quote.RolloverAmount > 0
	if quote.Eligible {
		sourceID := sub.ID
		quote.SourceSubscriptionID = &sourceID
		sourceExpiresAt := sub.ExpiresAt
		quote.SourceExpiresAt = &sourceExpiresAt
	}
	return quote
}

func clampPercent(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return value
}

func roundRenewalAmount(value float64) float64 {
	return math.Round(value*1e10) / 1e10
}

func renewalDiscountedAmount(amount, percent float64) float64 {
	return roundRenewalCurrency(amount * (1 - clampPercent(percent)/100))
}

func roundRenewalCurrency(value float64) float64 {
	return math.Round(value*100) / 100
}
