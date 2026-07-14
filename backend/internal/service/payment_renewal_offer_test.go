package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type renewalOfferUserSubRepo struct {
	userSubRepoNoop
	sub *UserSubscription
}

func (r renewalOfferUserSubRepo) GetByUserIDAndGroupID(context.Context, int64, int64) (*UserSubscription, error) {
	return r.sub, nil
}

func TestQuoteRenewalOfferWithinWindow(t *testing.T) {
	monthlyLimit := 100.0
	now := time.Now()
	sub := &UserSubscription{
		ID:              42,
		ExpiresAt:       now.Add(10 * 24 * time.Hour),
		MonthlyUsageUSD: 40,
		Group:           &Group{MonthlyLimitUSD: &monthlyLimit},
	}
	svc := &PaymentService{subscriptionSvc: &SubscriptionService{userSubRepo: renewalOfferUserSubRepo{sub: sub}}}

	quote := svc.quoteRenewalOffer(context.Background(), 1, 2, &PaymentConfig{
		RenewalOfferEnabled:    true,
		RenewalWindowDays:      14,
		RenewalDiscountPercent: 10,
		RenewalRolloverPercent: 20,
	}, now)

	require.True(t, quote.Eligible)
	require.Equal(t, 10.0, quote.DiscountPercent)
	require.Equal(t, 12.0, quote.RolloverAmount)
	require.NotNil(t, quote.SourceSubscriptionID)
	require.Equal(t, int64(42), *quote.SourceSubscriptionID)
	require.Equal(t, 90.0, renewalDiscountedAmount(100, quote.DiscountPercent))
}

func TestQuoteRenewalOfferOutsideWindow(t *testing.T) {
	now := time.Now()
	svc := &PaymentService{subscriptionSvc: &SubscriptionService{userSubRepo: renewalOfferUserSubRepo{sub: &UserSubscription{
		ID:        42,
		ExpiresAt: now.Add(-15 * 24 * time.Hour),
	}}}}

	quote := svc.quoteRenewalOffer(context.Background(), 1, 2, &PaymentConfig{
		RenewalOfferEnabled: true,
		RenewalWindowDays:   14,
	}, now)

	require.False(t, quote.Eligible)
	require.Zero(t, quote.RolloverAmount)
}

func TestUpdatePaymentConfigRejectsInvalidRenewalPercent(t *testing.T) {
	svc := &PaymentConfigService{settingRepo: &paymentConfigSettingRepoStub{values: map[string]string{}}}
	invalid := 101.0
	err := svc.UpdatePaymentConfig(context.Background(), UpdatePaymentConfigRequest{RenewalDiscountPercent: &invalid})
	require.Error(t, err)
}
