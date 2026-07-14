package service

import (
	"context"
	"math"
	"time"
)

type renewalOfferQuote struct {
	Eligible             bool
	DiscountPercent      float64
	RolloverAmount       float64
	SourceSubscriptionID *int64
	SourceExpiresAt      *time.Time
}

type RenewalOfferCheckoutSettings struct {
	Enabled         bool    `json:"enabled"`
	WindowDays      int     `json:"window_days"`
	RolloverPercent float64 `json:"rollover_percent"`
	DiscountPercent float64 `json:"discount_percent"`
}

type RenewalOfferQuote struct {
	Eligible         bool    `json:"eligible"`
	DiscountPercent  float64 `json:"discount_percent"`
	DiscountedAmount float64 `json:"discounted_amount"`
	RolloverAmount   float64 `json:"rollover_amount"`
}

func (s *PaymentService) QuoteRenewalOffer(ctx context.Context, userID, groupID int64, planPrice float64, cfg *PaymentConfig) RenewalOfferQuote {
	quote := s.quoteRenewalOffer(ctx, userID, groupID, cfg, time.Now())
	return RenewalOfferQuote{
		Eligible:         quote.Eligible,
		DiscountPercent:  quote.DiscountPercent,
		DiscountedAmount: renewalDiscountedAmount(planPrice, quote.DiscountPercent),
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
	settings.WindowDays = cfg.RenewalWindowDays
	settings.RolloverPercent = cfg.RenewalRolloverPercent
	settings.DiscountPercent = cfg.RenewalDiscountPercent
	return settings
}

func (s *PaymentService) quoteRenewalOffer(ctx context.Context, userID, groupID int64, cfg *PaymentConfig, now time.Time) renewalOfferQuote {
	quote := renewalOfferQuote{}
	if s == nil || cfg == nil || !cfg.RenewalOfferEnabled || cfg.RenewalWindowDays <= 0 || s.subscriptionSvc == nil || s.subscriptionSvc.userSubRepo == nil {
		return quote
	}
	sub, err := s.subscriptionSvc.userSubRepo.GetByUserIDAndGroupID(ctx, userID, groupID)
	if err != nil || sub == nil {
		return quote
	}
	window := time.Duration(cfg.RenewalWindowDays) * 24 * time.Hour
	if now.Before(sub.ExpiresAt.Add(-window)) || now.After(sub.ExpiresAt.Add(window)) {
		return quote
	}

	quote.Eligible = true
	quote.DiscountPercent = clampPercent(cfg.RenewalDiscountPercent)
	sourceID := sub.ID
	quote.SourceSubscriptionID = &sourceID
	sourceExpiresAt := sub.ExpiresAt
	quote.SourceExpiresAt = &sourceExpiresAt
	if sub.Group != nil && sub.Group.MonthlyLimitUSD != nil && *sub.Group.MonthlyLimitUSD > 0 {
		remaining := math.Max(*sub.Group.MonthlyLimitUSD-sub.MonthlyUsageUSD, 0)
		quote.RolloverAmount = roundRenewalAmount(remaining * clampPercent(cfg.RenewalRolloverPercent) / 100)
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
	return math.Round(amount*(1-clampPercent(percent)/100)*100) / 100
}
