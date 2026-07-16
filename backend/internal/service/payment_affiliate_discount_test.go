//go:build unit

package service

import (
	"context"
	"testing"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/stretchr/testify/require"
)

func TestQuoteAffiliateDiscountAppliesConfiguredCaps(t *testing.T) {
	ctx := context.Background()
	settingSvc := NewSettingService(&paymentFulfillmentSettingRepoStub{values: map[string]string{
		SettingKeyAffiliateEnabled:              "true",
		SettingKeyAffiliateDiscountEnabled:      "true",
		SettingKeyAffiliateDiscountMaxPercent:   "50",
		SettingKeyAffiliateDiscountMinPayAmount: "1",
	}}, nil)
	affiliateRepo := &paymentFulfillmentAffiliateRepoStub{availableDiscountQuota: 200}
	svc := &PaymentService{
		affiliateService: NewAffiliateService(affiliateRepo, settingSvc, nil, nil),
	}

	quote, err := svc.quoteAffiliateDiscount(ctx, CreateOrderRequest{UserID: 123}, 100)

	require.NoError(t, err)
	require.Equal(t, 100.0, quote.OriginalAmount)
	require.Equal(t, 50.0, quote.DiscountAmount)
	require.Equal(t, 50.0, quote.PayableAmount)
}

func TestQuoteAffiliateDiscountRespectsMinimumPayAmount(t *testing.T) {
	ctx := context.Background()
	settingSvc := NewSettingService(&paymentFulfillmentSettingRepoStub{values: map[string]string{
		SettingKeyAffiliateEnabled:              "true",
		SettingKeyAffiliateDiscountEnabled:      "true",
		SettingKeyAffiliateDiscountMaxPercent:   "100",
		SettingKeyAffiliateDiscountMinPayAmount: "1",
	}}, nil)
	affiliateRepo := &paymentFulfillmentAffiliateRepoStub{availableDiscountQuota: 200}
	svc := &PaymentService{
		affiliateService: NewAffiliateService(affiliateRepo, settingSvc, nil, nil),
	}

	quote, err := svc.quoteAffiliateDiscount(ctx, CreateOrderRequest{UserID: 123}, 100)

	require.NoError(t, err)
	require.Equal(t, 99.0, quote.DiscountAmount)
	require.Equal(t, 1.0, quote.PayableAmount)
}

func TestQuoteAffiliateDiscountCanBeOptedOutByUser(t *testing.T) {
	ctx := context.Background()
	useDiscount := false
	settingSvc := NewSettingService(&paymentFulfillmentSettingRepoStub{values: map[string]string{
		SettingKeyAffiliateEnabled:              "true",
		SettingKeyAffiliateDiscountEnabled:      "true",
		SettingKeyAffiliateDiscountMaxPercent:   "100",
		SettingKeyAffiliateDiscountMinPayAmount: "0",
	}}, nil)
	affiliateRepo := &paymentFulfillmentAffiliateRepoStub{availableDiscountQuota: 200}
	svc := &PaymentService{
		affiliateService: NewAffiliateService(affiliateRepo, settingSvc, nil, nil),
	}

	quote, err := svc.quoteAffiliateDiscount(ctx, CreateOrderRequest{
		UserID:               123,
		UseAffiliateDiscount: &useDiscount,
	}, 100)

	require.NoError(t, err)
	require.Zero(t, quote.DiscountAmount)
	require.Equal(t, 100.0, quote.PayableAmount)
}

func TestAffiliateDiscountRefundAmountUsesOrderAmountUnit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		order        *dbent.PaymentOrder
		refundAmount float64
		want         float64
	}{
		{
			name: "balance package refund is prorated by credited points",
			order: &dbent.PaymentOrder{
				OrderType:         payment.OrderTypeBalance,
				Amount:            45,
				OriginalAmount:    20,
				AffiliateDiscount: 4,
				PayAmount:         16,
			},
			refundAmount: 22.5,
			want:         2,
		},
		{
			name: "renewal full refund restores the full checkout discount",
			order: &dbent.PaymentOrder{
				OrderType:         payment.OrderTypeSubscription,
				Amount:            90,
				OriginalAmount:    100,
				AffiliateDiscount: 45,
				PayAmount:         45,
			},
			refundAmount: 90,
			want:         45,
		},
		{
			name: "renewal partial refund uses the discounted order base",
			order: &dbent.PaymentOrder{
				OrderType:         payment.OrderTypeSubscription,
				Amount:            90,
				OriginalAmount:    100,
				AffiliateDiscount: 45,
				PayAmount:         45,
			},
			refundAmount: 45,
			want:         22.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, affiliateDiscountRefundAmount(tt.order, tt.refundAmount))
		})
	}
}
