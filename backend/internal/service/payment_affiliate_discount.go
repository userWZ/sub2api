package service

import (
	"context"
	"math"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/shopspring/decimal"
)

type affiliateDiscountQuote struct {
	OriginalAmount float64
	DiscountAmount float64
	PayableAmount  float64
}

type AffiliateDiscountCheckoutSettings struct {
	Enabled      bool    `json:"enabled"`
	MaxPercent   float64 `json:"max_percent"`
	MinPayAmount float64 `json:"min_pay_amount"`
}

func (s *PaymentService) AffiliateDiscountCheckoutSettings(ctx context.Context) AffiliateDiscountCheckoutSettings {
	settings := AffiliateDiscountCheckoutSettings{
		Enabled:      false,
		MaxPercent:   AffiliateDiscountMaxPercentDefault,
		MinPayAmount: AffiliateDiscountMinPayAmountDefault,
	}
	if s == nil || s.affiliateService == nil || !s.affiliateService.IsEnabled(ctx) {
		return settings
	}
	settings.Enabled = true
	if s.affiliateService.settingService != nil {
		settings.Enabled = s.affiliateService.settingService.IsAffiliateDiscountEnabled(ctx)
		settings.MaxPercent = s.affiliateService.settingService.GetAffiliateDiscountMaxPercent(ctx)
		settings.MinPayAmount = s.affiliateService.settingService.GetAffiliateDiscountMinPayAmount(ctx)
	}
	return settings
}

func (s *PaymentService) quoteAffiliateDiscount(ctx context.Context, req CreateOrderRequest, originalAmount float64) (affiliateDiscountQuote, error) {
	quote := affiliateDiscountQuote{
		OriginalAmount: roundMoney(originalAmount),
		PayableAmount:  roundMoney(originalAmount),
	}
	if originalAmount <= 0 || s == nil || s.affiliateService == nil {
		return quote, nil
	}
	if req.UseAffiliateDiscount != nil && !*req.UseAffiliateDiscount {
		return quote, nil
	}
	if !s.affiliateService.IsEnabled(ctx) {
		return quote, nil
	}
	if s.affiliateService.settingService != nil && !s.affiliateService.settingService.IsAffiliateDiscountEnabled(ctx) {
		return quote, nil
	}

	available, err := s.affiliateService.GetAvailableDiscountQuota(ctx, req.UserID)
	if err != nil {
		return quote, err
	}
	if available <= 0 {
		return quote, nil
	}

	maxPercent := AffiliateDiscountMaxPercentDefault
	minPayAmount := AffiliateDiscountMinPayAmountDefault
	if s.affiliateService.settingService != nil {
		maxPercent = s.affiliateService.settingService.GetAffiliateDiscountMaxPercent(ctx)
		minPayAmount = s.affiliateService.settingService.GetAffiliateDiscountMinPayAmount(ctx)
	}

	maxByPercent := decimal.NewFromFloat(originalAmount).
		Mul(decimal.NewFromFloat(maxPercent)).
		Div(decimal.NewFromInt(100))
	maxByMinPay := decimal.NewFromFloat(originalAmount).Sub(decimal.NewFromFloat(minPayAmount))
	if maxByMinPay.IsNegative() {
		maxByMinPay = decimal.Zero
	}

	discount := decimal.Min(
		decimal.NewFromFloat(available),
		decimal.Min(maxByPercent, maxByMinPay),
	).Round(2)
	if !discount.IsPositive() {
		return quote, nil
	}
	payable := decimal.NewFromFloat(originalAmount).Sub(discount).Round(2)
	if payable.LessThan(decimal.Zero) {
		payable = decimal.Zero
	}

	quote.DiscountAmount = discount.InexactFloat64()
	quote.PayableAmount = payable.InexactFloat64()
	return quote, nil
}

func roundMoney(v float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0
	}
	return decimal.NewFromFloat(v).Round(2).InexactFloat64()
}

func affiliateDiscountRefundAmount(order *dbent.PaymentOrder, refundAmount float64) float64 {
	if order == nil || order.AffiliateDiscount <= 0 || refundAmount <= 0 {
		return 0
	}
	// refundAmount uses the same unit as order.Amount: points for balance orders
	// and payment currency for subscription orders. Using OriginalAmount here
	// mixes CNY with points for balance packages and also breaks renewed plans
	// whose discounted order amount differs from the list price.
	base := order.Amount
	if base <= 0 {
		return 0
	}
	if refundAmount >= base-paymentAmountToleranceForCurrency(PaymentOrderCurrency(order)) {
		return roundMoney(order.AffiliateDiscount)
	}
	return decimal.NewFromFloat(order.AffiliateDiscount).
		Mul(decimal.NewFromFloat(refundAmount)).
		Div(decimal.NewFromFloat(base)).
		Round(2).
		InexactFloat64()
}

func paymentOrderOriginalAmount(order *dbent.PaymentOrder) float64 {
	if order == nil {
		return 0
	}
	if order.OriginalAmount > 0 {
		return order.OriginalAmount
	}
	if order.PayAmount > 0 {
		return order.PayAmount
	}
	return order.Amount
}

func PaymentOrderOriginalAmountForResponse(order *dbent.PaymentOrder) float64 {
	return paymentOrderOriginalAmount(order)
}

func (s *PaymentService) restoreAffiliateDiscountForOrder(ctx context.Context, order *dbent.PaymentOrder, reason string) error {
	if s == nil || s.affiliateService == nil || order == nil || order.AffiliateDiscount <= 0 {
		return nil
	}
	restored, err := s.affiliateService.RestoreDiscountForOrder(ctx, order.UserID, order.AffiliateDiscount, order.ID)
	if err != nil {
		s.writeAuditLog(ctx, order.ID, "AFFILIATE_DISCOUNT_RESTORE_FAILED", "system", map[string]any{
			"reason": reason,
			"amount": order.AffiliateDiscount,
			"error":  err.Error(),
		})
		return err
	}
	if restored {
		s.writeAuditLog(ctx, order.ID, "AFFILIATE_DISCOUNT_RESTORED", "system", map[string]any{
			"reason": reason,
			"amount": order.AffiliateDiscount,
		})
	}
	return nil
}

func (s *PaymentService) restoreAffiliateDiscountForRefund(ctx context.Context, order *dbent.PaymentOrder, refundAmount float64) error {
	discountRefund := affiliateDiscountRefundAmount(order, refundAmount)
	if s == nil || s.affiliateService == nil || order == nil || discountRefund <= 0 {
		return nil
	}
	restored, err := s.affiliateService.RestoreDiscountForOrder(ctx, order.UserID, discountRefund, order.ID)
	if err != nil {
		s.writeAuditLog(ctx, order.ID, "AFFILIATE_DISCOUNT_REFUND_FAILED", "system", map[string]any{
			"refundAmount":      refundAmount,
			"discountRefund":    discountRefund,
			"affiliateDiscount": order.AffiliateDiscount,
			"error":             err.Error(),
		})
		return err
	}
	if restored {
		s.writeAuditLog(ctx, order.ID, "AFFILIATE_DISCOUNT_REFUNDED", "system", map[string]any{
			"refundAmount":      refundAmount,
			"discountRefund":    discountRefund,
			"affiliateDiscount": order.AffiliateDiscount,
		})
	}
	return nil
}

func (s *PaymentService) reverseAffiliateRebateForFullRefund(ctx context.Context, order *dbent.PaymentOrder, refundAmount float64) error {
	if s == nil || s.affiliateService == nil || order == nil || refundAmount <= 0 {
		return nil
	}
	if refundAmount < order.Amount-paymentAmountToleranceForCurrency(PaymentOrderCurrency(order)) {
		return nil
	}
	reversed, err := s.affiliateService.ReverseAccruedRebateForOrder(ctx, order.ID)
	if err != nil {
		s.writeAuditLog(ctx, order.ID, "AFFILIATE_REBATE_REVERSAL_FAILED", "system", map[string]any{
			"refundAmount": refundAmount,
			"error":        err.Error(),
		})
		return err
	}
	if reversed > 0 {
		s.writeAuditLog(ctx, order.ID, "AFFILIATE_REBATE_REVERSED", "system", map[string]any{
			"refundAmount":   refundAmount,
			"reversedAmount": reversed,
		})
	}
	return nil
}
