//go:build unit

package service

import (
	"context"
	"encoding/json"
	"testing"
)

func TestResolvePublicHomePricingConfigUsesDefaultWhenUnset(t *testing.T) {
	svc := &PaymentConfigService{
		settingRepo: &settingRepoStub{values: map[string]string{}},
	}

	raw := svc.ResolvePublicHomePricingConfig(context.Background(), "")
	if len(raw) == 0 {
		t.Fatal("ResolvePublicHomePricingConfig returned empty config")
	}

	var cfg HomePricingPublicConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		t.Fatalf("unmarshal public config: %v", err)
	}
	if len(cfg.CreditCards) == 0 {
		t.Fatal("CreditCards is empty, want default credit packages")
	}
	if cfg.CreditCards[0].RechargeAmount != 20 || cfg.CreditCards[0].CreditedAmount != 80 {
		t.Fatalf("first default credit card = %#v, want ¥20 for 80 credits", cfg.CreditCards[0])
	}
}
