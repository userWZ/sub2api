//go:build unit

package service

import (
	"context"
	"testing"

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
