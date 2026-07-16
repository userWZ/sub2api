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

	quote := svc.quoteRenewalOffer(context.Background(), 1, 2, 100, &PaymentConfig{
		RenewalOfferEnabled:    true,
		RenewalBeforeDays:      14,
		RenewalAfterDays:       14,
		RenewalDiscountEnabled: true,
		RenewalDiscountPercent: 10,
		RenewalRolloverEnabled: true,
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

	quote := svc.quoteRenewalOffer(context.Background(), 1, 2, 100, &PaymentConfig{
		RenewalOfferEnabled: true,
		RenewalBeforeDays:   14,
		RenewalAfterDays:    14,
	}, now)

	require.False(t, quote.Eligible)
	require.Zero(t, quote.RolloverAmount)
}

func TestQuoteRenewalOfferUsesIndependentSwitchesThresholdsAndCaps(t *testing.T) {
	monthlyLimit := 100.0
	now := time.Now()
	sub := &UserSubscription{
		ID:              43,
		ExpiresAt:       now.Add(2 * 24 * time.Hour),
		MonthlyUsageUSD: 10,
		Group:           &Group{MonthlyLimitUSD: &monthlyLimit},
	}
	svc := &PaymentService{subscriptionSvc: &SubscriptionService{userSubRepo: renewalOfferUserSubRepo{sub: sub}}}

	quote := svc.quoteRenewalOffer(context.Background(), 1, 2, 200, &PaymentConfig{
		RenewalOfferEnabled:      true,
		RenewalBeforeDays:        7,
		RenewalAfterDays:         3,
		RenewalDiscountEnabled:   true,
		RenewalDiscountPercent:   10,
		RenewalDiscountMinOrder:  100,
		RenewalDiscountMaxAmount: 12,
		RenewalRolloverEnabled:   true,
		RenewalRolloverPercent:   20,
		RenewalRolloverMinUnused: 50,
		RenewalRolloverMaxAmount: 15,
	}, now)

	require.True(t, quote.Eligible)
	require.Equal(t, 12.0, quote.DiscountAmount)
	require.Equal(t, 15.0, quote.RolloverAmount)
}

func TestQuoteRenewalOfferAllowsRolloverWithoutDiscount(t *testing.T) {
	monthlyLimit := 100.0
	now := time.Now()
	svc := &PaymentService{subscriptionSvc: &SubscriptionService{userSubRepo: renewalOfferUserSubRepo{sub: &UserSubscription{
		ID: 44, ExpiresAt: now.Add(-2 * 24 * time.Hour), MonthlyUsageUSD: 50,
		Group: &Group{MonthlyLimitUSD: &monthlyLimit},
	}}}}

	quote := svc.quoteRenewalOffer(context.Background(), 1, 2, 100, &PaymentConfig{
		RenewalOfferEnabled:    true,
		RenewalBeforeDays:      0,
		RenewalAfterDays:       3,
		RenewalDiscountEnabled: false,
		RenewalRolloverEnabled: true,
		RenewalRolloverPercent: 20,
	}, now)

	require.True(t, quote.Eligible)
	require.Zero(t, quote.DiscountAmount)
	require.Equal(t, 10.0, quote.RolloverAmount)
}

func TestUpdateRenewalSettingsPersistsNormalizedDedicatedConfig(t *testing.T) {
	repo := &paymentConfigSettingRepoStub{values: map[string]string{}}
	svc := &PaymentConfigService{settingRepo: repo}

	updated, err := svc.UpdateRenewalSettings(context.Background(), RenewalSettings{
		OfferEnabled: true, BeforeExpiryDays: 14, AfterExpiryDays: 7,
		DiscountEnabled: true, DiscountPercent: 10, DiscountMinOrder: 5, DiscountMaxAmount: 20,
		RolloverEnabled: true, RolloverPercent: 20, RolloverMinUnused: 1, RolloverMaxAmount: 30,
		EmailEnabled: true, EmailReminderDays: []int{1, 14, 7, 14},
	})

	require.NoError(t, err)
	require.Equal(t, []int{14, 7, 1}, updated.EmailReminderDays)
	require.Equal(t, "true", repo.values[SettingRenewalDiscountEnabled])
	require.Equal(t, "true", repo.values[SettingRenewalRolloverEnabled])
	require.Equal(t, "14,7,1", repo.values[SettingRenewalEmailReminderDays])
	require.Equal(t, "14", repo.values[SettingRenewalWindowDays])
}

func TestUpdateRenewalSettingsRejectsReminderOutsideWindow(t *testing.T) {
	svc := &PaymentConfigService{settingRepo: &paymentConfigSettingRepoStub{values: map[string]string{}}}
	_, err := svc.UpdateRenewalSettings(context.Background(), RenewalSettings{
		BeforeExpiryDays:  7,
		DiscountPercent:   10,
		RolloverPercent:   20,
		EmailEnabled:      true,
		EmailReminderDays: []int{14},
	})
	require.Error(t, err)
}

func TestUpdatePaymentConfigRejectsInvalidRenewalPercent(t *testing.T) {
	svc := &PaymentConfigService{settingRepo: &paymentConfigSettingRepoStub{values: map[string]string{}}}
	invalid := 101.0
	err := svc.UpdatePaymentConfig(context.Background(), UpdatePaymentConfigRequest{RenewalDiscountPercent: &invalid})
	require.Error(t, err)
}
