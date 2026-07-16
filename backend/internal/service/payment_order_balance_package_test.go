//go:build unit

package service

import (
	"context"
	"encoding/json"
	"testing"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

func TestResolveBalanceOrderPackageAmountUsesConfiguredCredit(t *testing.T) {
	cfg := defaultHomePricingConfig()
	cfg.CreditCards = []HomePricingCreditCardConfig{
		{
			ID:             "credit-80",
			Enabled:        true,
			SortOrder:      10,
			RechargeAmount: 20,
			CreditedAmount: 80,
			Name:           HomePricingLocalizedText{Zh: "80 积分", En: "80 Credits"},
			Description:    HomePricingLocalizedText{Zh: "20 元购买 80 积分", En: "Pay ¥20 for 80 credits"},
		},
	}
	raw, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}
	svc := &PaymentService{
		configService: &PaymentConfigService{
			settingRepo: &settingRepoStub{values: map[string]string{
				SettingKeyHomePricingConfig: string(raw),
			}},
		},
	}

	credited, err := svc.resolveBalanceOrderPackageAmount(context.Background(), 20, &PaymentConfig{BalanceRechargeMultiplier: 1})
	if err != nil {
		t.Fatalf("resolveBalanceOrderPackageAmount returned error: %v", err)
	}
	if credited != 80 {
		t.Fatalf("credited amount = %v, want 80", credited)
	}
}

func TestResolveBalanceOrderPackageAmountRejectsArbitraryAmount(t *testing.T) {
	svc := &PaymentService{
		configService: &PaymentConfigService{
			settingRepo: &settingRepoStub{values: map[string]string{}},
		},
	}

	_, err := svc.resolveBalanceOrderPackageAmount(context.Background(), 12.34, &PaymentConfig{BalanceRechargeMultiplier: 1})
	if err == nil {
		t.Fatal("resolveBalanceOrderPackageAmount returned nil error")
	}
	appErr := infraerrors.FromError(err)
	if appErr == nil || appErr.Reason != "BALANCE_PACKAGE_NOT_AVAILABLE" {
		t.Fatalf("reason = %#v, want BALANCE_PACKAGE_NOT_AVAILABLE", appErr)
	}
}
