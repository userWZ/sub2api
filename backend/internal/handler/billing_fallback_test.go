package handler

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestCanFallbackSubscriptionBillingToBalance(t *testing.T) {
	subscriptionGroup := &service.Group{SubscriptionType: service.SubscriptionTypeSubscription}
	standardGroup := &service.Group{SubscriptionType: service.SubscriptionTypeStandard}
	subscription := &service.UserSubscription{}

	tests := []struct {
		name         string
		err          error
		apiKey       *service.APIKey
		subscription *service.UserSubscription
		want         bool
	}{
		{
			name:         "daily subscription limit falls back when balance fallback is enabled",
			err:          service.ErrDailyLimitExceeded,
			apiKey:       &service.APIKey{QuotaDisabled: true, Group: subscriptionGroup},
			subscription: subscription,
			want:         true,
		},
		{
			name:         "invalid subscription falls back when balance fallback is enabled",
			err:          service.ErrSubscriptionInvalid,
			apiKey:       &service.APIKey{QuotaDisabled: true, Group: subscriptionGroup},
			subscription: subscription,
			want:         true,
		},
		{
			name:         "balance errors do not trigger another fallback",
			err:          service.ErrInsufficientBalance,
			apiKey:       &service.APIKey{QuotaDisabled: true, Group: subscriptionGroup},
			subscription: subscription,
			want:         false,
		},
		{
			name:         "key must enable balance fallback",
			err:          service.ErrDailyLimitExceeded,
			apiKey:       &service.APIKey{QuotaDisabled: false, Group: subscriptionGroup},
			subscription: subscription,
			want:         false,
		},
		{
			name:         "standard groups are already balance mode",
			err:          service.ErrDailyLimitExceeded,
			apiKey:       &service.APIKey{QuotaDisabled: true, Group: standardGroup},
			subscription: subscription,
			want:         false,
		},
		{
			name:         "requires a subscription context",
			err:          service.ErrDailyLimitExceeded,
			apiKey:       &service.APIKey{QuotaDisabled: true, Group: subscriptionGroup},
			subscription: nil,
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, canFallbackSubscriptionBillingToBalance(tt.err, tt.apiKey, tt.subscription))
		})
	}
}
