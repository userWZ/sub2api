package handler

import (
	"context"
	"errors"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

func checkBillingEligibilityWithBalanceFallback(
	ctx context.Context,
	billingCacheService *service.BillingCacheService,
	apiKeyService *service.APIKeyService,
	subscriptionService *service.SubscriptionService,
	apiKey *service.APIKey,
	subscription *service.UserSubscription,
	platform string,
) (*service.UserSubscription, error) {
	if billingCacheService == nil || apiKey == nil || apiKey.User == nil {
		return subscription, nil
	}
	if strings.TrimSpace(platform) == "" {
		platform = service.QuotaPlatform(ctx, apiKey)
	}

	err := billingCacheService.CheckBillingEligibility(ctx, apiKey.User, apiKey, apiKey.Group, subscription, platform)
	if err == nil {
		return subscription, nil
	}
	if !canFallbackSubscriptionBillingToBalance(err, apiKey, subscription) {
		return subscription, err
	}
	if alternative, altErr := tryAlternativeAutomaticSubscription(ctx, billingCacheService, subscriptionService, apiKey, subscription, platform); altErr != nil {
		return subscription, altErr
	} else if alternative != nil {
		return alternative, nil
	}
	if apiKeyService == nil {
		return subscription, err
	}

	req, _ := ctx.Value(ctxkey.SubscriptionEntitlementRequest).(service.EntitlementRequest)
	group, groupErr := apiKeyService.ResolveDefaultStandardGroupForRequest(ctx, apiKey.User.ID, req)
	if groupErr != nil {
		return subscription, groupErr
	}
	if group == nil {
		return subscription, err
	}

	previousGroupID := apiKey.GroupID
	previousGroup := apiKey.Group
	groupID := group.ID
	apiKey.GroupID = &groupID
	apiKey.Group = group

	if balanceErr := billingCacheService.CheckBillingEligibility(ctx, apiKey.User, apiKey, group, nil, platform); balanceErr != nil {
		apiKey.GroupID = previousGroupID
		apiKey.Group = previousGroup
		return subscription, balanceErr
	}

	return nil, nil
}

func tryAlternativeAutomaticSubscription(
	ctx context.Context,
	billingCacheService *service.BillingCacheService,
	subscriptionService *service.SubscriptionService,
	apiKey *service.APIKey,
	current *service.UserSubscription,
	platform string,
) (*service.UserSubscription, error) {
	if ctx == nil || billingCacheService == nil || subscriptionService == nil || apiKey == nil || apiKey.User == nil || current == nil {
		return nil, nil
	}
	if automatic, _ := ctx.Value(ctxkey.AutomaticSubscriptionResolved).(bool); !automatic {
		return nil, nil
	}

	req, _ := ctx.Value(ctxkey.SubscriptionEntitlementRequest).(service.EntitlementRequest)
	candidates, err := subscriptionService.ResolveAutomaticSubscriptionCandidatesForRequest(ctx, apiKey.User.ID, req)
	if err != nil {
		return nil, err
	}

	previousGroupID := apiKey.GroupID
	previousGroup := apiKey.Group
	for i := range candidates {
		candidate := candidates[i]
		if candidate.Subscription.ID == current.ID || candidate.Group == nil {
			continue
		}

		groupID := candidate.Group.ID
		apiKey.GroupID = &groupID
		apiKey.Group = candidate.Group
		candidate.Subscription.Group = candidate.Group
		altSub := &candidate.Subscription
		if err := billingCacheService.CheckBillingEligibility(ctx, apiKey.User, apiKey, candidate.Group, altSub, platform); err == nil {
			return altSub, nil
		} else if canFallbackSubscriptionBillingToBalance(err, apiKey, altSub) {
			continue
		} else {
			apiKey.GroupID = previousGroupID
			apiKey.Group = previousGroup
			return nil, err
		}
	}
	apiKey.GroupID = previousGroupID
	apiKey.Group = previousGroup
	return nil, nil
}

func canFallbackSubscriptionBillingToBalance(err error, apiKey *service.APIKey, subscription *service.UserSubscription) bool {
	if err == nil || apiKey == nil || subscription == nil {
		return false
	}
	if !apiKey.QuotaDisabled {
		return false
	}
	if apiKey.Group == nil || !apiKey.Group.IsSubscriptionType() {
		return false
	}
	return errors.Is(err, service.ErrDailyLimitExceeded) ||
		errors.Is(err, service.ErrWeeklyLimitExceeded) ||
		errors.Is(err, service.ErrMonthlyLimitExceeded) ||
		errors.Is(err, service.ErrSubscriptionInvalid)
}
