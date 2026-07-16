//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type affiliateUsageRewardRepoStub struct {
	*paymentFulfillmentAffiliateRepoStub
	processCalls int
	lastEvent    AffiliateUsageRewardEvent
	lastSettings AffiliateSettings
	result       *AffiliateUsageRewardProcessResult
}

func (s *affiliateUsageRewardRepoStub) ProcessUsageReward(_ context.Context, event AffiliateUsageRewardEvent, settings AffiliateSettings) (*AffiliateUsageRewardProcessResult, error) {
	s.processCalls++
	s.lastEvent = event
	s.lastSettings = settings
	return s.result, nil
}

func (s *affiliateUsageRewardRepoStub) ListUsageRewardRecords(context.Context, AffiliateUsageRewardFilter) ([]AffiliateUsageRewardRecord, int64, error) {
	return nil, 0, nil
}

func (s *affiliateUsageRewardRepoStub) ReviewUsageReward(context.Context, int64, int64, bool, string) (*AffiliateUsageRewardReviewResult, error) {
	return nil, nil
}

func TestProcessFirstUsageReward_UsesConfiguredCreditsAndIsDisabledByDefault(t *testing.T) {
	ctx := context.Background()
	settingsRepo := &paymentFulfillmentSettingRepoStub{values: map[string]string{
		SettingKeyAffiliateEnabled:            "true",
		SettingKeyAffiliateUsageRewardEnabled: "false",
		SettingKeyAffiliateUsageRewardAmount:  "20",
	}}
	repo := &affiliateUsageRewardRepoStub{
		paymentFulfillmentAffiliateRepoStub: &paymentFulfillmentAffiliateRepoStub{},
		result:                              &AffiliateUsageRewardProcessResult{Created: true, InviterID: 9, Status: "granted"},
	}
	svc := NewAffiliateService(repo, NewSettingService(settingsRepo, nil), nil, nil)

	result, err := svc.ProcessFirstUsageReward(ctx, AffiliateUsageRewardEvent{InviteeUserID: 10, ActualCost: 0.01})
	require.NoError(t, err)
	require.Nil(t, result)
	require.Zero(t, repo.processCalls)

	settingsRepo.values[SettingKeyAffiliateUsageRewardEnabled] = "true"
	result, err = svc.ProcessFirstUsageReward(ctx, AffiliateUsageRewardEvent{
		InviteeUserID: 10,
		RequestID:     "req-1",
		APIKeyID:      11,
		IPAddress:     "203.0.113.10",
		ActualCost:    0.01,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, 1, repo.processCalls)
	require.Equal(t, 20.0, repo.lastSettings.UsageRewardAmount)
	require.Equal(t, int64(10), repo.lastEvent.InviteeUserID)
}

func TestAffiliateSettings_DefaultRiskLimitsAndValidation(t *testing.T) {
	settingsRepo := &paymentFulfillmentSettingRepoStub{values: map[string]string{}}
	svc := NewAffiliateService(
		&affiliateUsageRewardRepoStub{paymentFulfillmentAffiliateRepoStub: &paymentFulfillmentAffiliateRepoStub{}},
		NewSettingService(settingsRepo, nil),
		nil,
		nil,
	)

	settings, err := svc.GetSettings(context.Background())
	require.NoError(t, err)
	require.Equal(t, 20.0, settings.UsageRewardAmount)
	require.Equal(t, 5, settings.RewardInviterDailyLimit)
	require.Equal(t, 30, settings.RewardInviterRolling30DayLimit)
	require.Equal(t, 2, settings.RewardIPDailyLimit)
	require.True(t, settings.RewardReviewSameIP)
	require.True(t, settings.RewardReviewMissingIP)

	settings.UsageRewardAmount = 0
	_, err = svc.UpdateSettings(context.Background(), *settings)
	require.Error(t, err)
}
