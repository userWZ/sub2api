package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/subscriptionplan"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const SettingKeyHomePricingConfig = "home_pricing_config"

type HomePricingLocalizedText struct {
	Zh string `json:"zh"`
	En string `json:"en"`
}

type HomePricingMetricConfig struct {
	Label HomePricingLocalizedText `json:"label"`
	Value HomePricingLocalizedText `json:"value"`
}

type HomePricingSubscriptionCardConfig struct {
	ID                 string                    `json:"id"`
	Enabled            bool                      `json:"enabled"`
	SortOrder          int                       `json:"sort_order"`
	SubscriptionPlanID int64                     `json:"subscription_plan_id"`
	Name               HomePricingLocalizedText  `json:"name"`
	Description        HomePricingLocalizedText  `json:"description"`
	Badge              HomePricingLocalizedText  `json:"badge"`
	Period             HomePricingLocalizedText  `json:"period"`
	Highlight          bool                      `json:"highlight"`
	Metrics            []HomePricingMetricConfig `json:"metrics"`
}

type HomePricingCreditCardConfig struct {
	ID             string                    `json:"id"`
	Enabled        bool                      `json:"enabled"`
	SortOrder      int                       `json:"sort_order"`
	RechargeAmount float64                   `json:"recharge_amount"`
	CreditedAmount float64                   `json:"credited_amount"`
	Name           HomePricingLocalizedText  `json:"name"`
	Description    HomePricingLocalizedText  `json:"description"`
	Badge          HomePricingLocalizedText  `json:"badge"`
	Period         HomePricingLocalizedText  `json:"period"`
	Highlight      bool                      `json:"highlight"`
	Metrics        []HomePricingMetricConfig `json:"metrics"`
}

type HomePricingGroupConfig struct {
	Title       HomePricingLocalizedText `json:"title"`
	Description HomePricingLocalizedText `json:"description"`
}

type HomePricingConfig struct {
	Eyebrow           HomePricingLocalizedText            `json:"eyebrow"`
	Title             HomePricingLocalizedText            `json:"title"`
	Description       HomePricingLocalizedText            `json:"description"`
	SubscriptionGroup HomePricingGroupConfig              `json:"subscription_group"`
	CreditGroup       HomePricingGroupConfig              `json:"credit_group"`
	SubscriptionCards []HomePricingSubscriptionCardConfig `json:"subscription_cards"`
	CreditCards       []HomePricingCreditCardConfig       `json:"credit_cards"`
}

type HomePricingPublicConfig struct {
	Eyebrow           HomePricingLocalizedText      `json:"eyebrow"`
	Title             HomePricingLocalizedText      `json:"title"`
	Description       HomePricingLocalizedText      `json:"description"`
	SubscriptionGroup HomePricingGroupConfig        `json:"subscription_group"`
	CreditGroup       HomePricingGroupConfig        `json:"credit_group"`
	SubscriptionCards []HomePricingPublicPlanCard   `json:"subscription_cards"`
	CreditCards       []HomePricingPublicCreditCard `json:"credit_cards"`
}

type HomePricingPublicPlanCard struct {
	ID                 string                    `json:"id"`
	Enabled            bool                      `json:"enabled"`
	SortOrder          int                       `json:"sort_order"`
	SubscriptionPlanID int64                     `json:"subscription_plan_id"`
	Name               HomePricingLocalizedText  `json:"name"`
	Description        HomePricingLocalizedText  `json:"description"`
	Badge              HomePricingLocalizedText  `json:"badge"`
	Period             HomePricingLocalizedText  `json:"period"`
	Highlight          bool                      `json:"highlight"`
	Metrics            []HomePricingMetricConfig `json:"metrics"`
	Price              float64                   `json:"price"`
	OriginalPrice      *float64                  `json:"original_price,omitempty"`
	ForSale            bool                      `json:"for_sale"`
}

type HomePricingPublicCreditCard struct {
	ID             string                    `json:"id"`
	Enabled        bool                      `json:"enabled"`
	SortOrder      int                       `json:"sort_order"`
	RechargeAmount float64                   `json:"recharge_amount"`
	CreditedAmount float64                   `json:"credited_amount"`
	Name           HomePricingLocalizedText  `json:"name"`
	Description    HomePricingLocalizedText  `json:"description"`
	Badge          HomePricingLocalizedText  `json:"badge"`
	Period         HomePricingLocalizedText  `json:"period"`
	Highlight      bool                      `json:"highlight"`
	Metrics        []HomePricingMetricConfig `json:"metrics"`
	Price          float64                   `json:"price"`
}

func defaultHomePricingConfig() *HomePricingConfig {
	return &HomePricingConfig{
		Eyebrow: HomePricingLocalizedText{Zh: "定价", En: "Pricing"},
		Title: HomePricingLocalizedText{
			Zh: "选择适合你的使用方式。",
			En: "Choose the usage option that fits you.",
		},
		Description: HomePricingLocalizedText{
			Zh: "可以购买订阅获得周期服务额度，也可以购买余额积分包作为补充；实际消耗会随模型、输入输出长度和工具行为变化。",
			En: "Buy a subscription for recurring service credit, or add balance credit packages as backup. Actual usage varies by model, input/output length, and tool behavior.",
		},
		SubscriptionGroup: HomePricingGroupConfig{
			Title:       HomePricingLocalizedText{Zh: "订阅套餐", En: "Subscription Plans"},
			Description: HomePricingLocalizedText{Zh: "适合每天持续使用 AI 工具的用户。", En: "For users who run AI tools continuously."},
		},
		CreditGroup: HomePricingGroupConfig{
			Title:       HomePricingLocalizedText{Zh: "额度套餐", En: "Credit Packages"},
			Description: HomePricingLocalizedText{Zh: "适合月卡不够用时灵活补充，额度无每日限制。", En: "Flexible top-ups when your monthly plan is not enough. Credits have no daily limit."},
		},
		SubscriptionCards: []HomePricingSubscriptionCardConfig{},
		CreditCards: []HomePricingCreditCardConfig{
			{
				ID:             "credit-80",
				Enabled:        true,
				SortOrder:      10,
				RechargeAmount: 20,
				CreditedAmount: 80,
				Name:           HomePricingLocalizedText{Zh: "80 积分", En: "80 Credits"},
				Description:    HomePricingLocalizedText{Zh: "20 元购买 80 积分。", En: "Pay ¥20 for 80 credits."},
				Metrics: []HomePricingMetricConfig{
					{Label: HomePricingLocalizedText{Zh: "到账积分", En: "Credits received"}, Value: HomePricingLocalizedText{Zh: "80 积分", En: "80 credits"}},
					{Label: HomePricingLocalizedText{Zh: "类型", En: "Type"}, Value: HomePricingLocalizedText{Zh: "余额额度", En: "balance credit"}},
				},
			},
			{
				ID:             "credit-180",
				Enabled:        true,
				SortOrder:      20,
				RechargeAmount: 40,
				CreditedAmount: 180,
				Name:           HomePricingLocalizedText{Zh: "180 积分", En: "180 Credits"},
				Description:    HomePricingLocalizedText{Zh: "40 元购买 180 积分。", En: "Pay ¥40 for 180 credits."},
				Badge:          HomePricingLocalizedText{Zh: "常用", En: "Common"},
				Highlight:      true,
				Metrics: []HomePricingMetricConfig{
					{Label: HomePricingLocalizedText{Zh: "到账积分", En: "Credits received"}, Value: HomePricingLocalizedText{Zh: "180 积分", En: "180 credits"}},
					{Label: HomePricingLocalizedText{Zh: "类型", En: "Type"}, Value: HomePricingLocalizedText{Zh: "余额额度", En: "balance credit"}},
				},
			},
			{
				ID:             "credit-1000",
				Enabled:        true,
				SortOrder:      30,
				RechargeAmount: 200,
				CreditedAmount: 1000,
				Name:           HomePricingLocalizedText{Zh: "1000 积分", En: "1000 Credits"},
				Description:    HomePricingLocalizedText{Zh: "200 元购买 1000 积分。", En: "Pay ¥200 for 1000 credits."},
				Metrics: []HomePricingMetricConfig{
					{Label: HomePricingLocalizedText{Zh: "到账积分", En: "Credits received"}, Value: HomePricingLocalizedText{Zh: "1000 积分", En: "1000 credits"}},
					{Label: HomePricingLocalizedText{Zh: "类型", En: "Type"}, Value: HomePricingLocalizedText{Zh: "余额额度", En: "balance credit"}},
				},
			},
		},
	}
}

func parseHomePricingConfig(raw string) (*HomePricingConfig, error) {
	if strings.TrimSpace(raw) == "" {
		return defaultHomePricingConfig(), nil
	}
	var cfg HomePricingConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return nil, err
	}
	normalizeHomePricingConfig(&cfg)
	return &cfg, nil
}

func normalizeHomePricingConfig(cfg *HomePricingConfig) {
	if cfg.SubscriptionCards == nil {
		cfg.SubscriptionCards = []HomePricingSubscriptionCardConfig{}
	}
	if cfg.CreditCards == nil {
		cfg.CreditCards = []HomePricingCreditCardConfig{}
	}
}

func (s *PaymentConfigService) GetHomePricingConfig(ctx context.Context) (*HomePricingConfig, error) {
	raw, err := s.settingRepo.GetValue(ctx, SettingKeyHomePricingConfig)
	if err != nil && !errors.Is(err, ErrSettingNotFound) {
		return nil, fmt.Errorf("get home pricing config: %w", err)
	}
	cfg, err := parseHomePricingConfig(raw)
	if err != nil {
		return nil, infraerrors.InternalServer("HOME_PRICING_CONFIG_INVALID", "home pricing config is invalid").WithCause(err)
	}
	return cfg, nil
}

func (s *PaymentConfigService) UpdateHomePricingConfig(ctx context.Context, cfg HomePricingConfig) (*HomePricingConfig, error) {
	normalizeHomePricingConfig(&cfg)
	if err := s.validateHomePricingConfig(ctx, &cfg); err != nil {
		return nil, err
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("marshal home pricing config: %w", err)
	}
	if err := s.settingRepo.Set(ctx, SettingKeyHomePricingConfig, string(data)); err != nil {
		return nil, fmt.Errorf("save home pricing config: %w", err)
	}
	return &cfg, nil
}

func (s *PaymentConfigService) ResolvePublicHomePricingConfig(ctx context.Context, raw string) json.RawMessage {
	cfg, err := parseHomePricingConfig(raw)
	if err != nil {
		return nil
	}
	publicCfg := s.resolveHomePricingConfig(ctx, cfg, s.homePricingBalanceRechargeMultiplier(ctx))
	if len(publicCfg.SubscriptionCards) == 0 && len(publicCfg.CreditCards) == 0 {
		return nil
	}
	data, err := json.Marshal(publicCfg)
	if err != nil {
		return nil
	}
	return data
}

func (s *PaymentConfigService) resolveHomePricingConfig(ctx context.Context, cfg *HomePricingConfig, balanceRechargeMultiplier float64) HomePricingPublicConfig {
	plansByID := s.homePricingPlansByID(ctx, cfg.SubscriptionCards)
	out := HomePricingPublicConfig{
		Eyebrow:           cfg.Eyebrow,
		Title:             cfg.Title,
		Description:       cfg.Description,
		SubscriptionGroup: cfg.SubscriptionGroup,
		CreditGroup:       cfg.CreditGroup,
		SubscriptionCards: []HomePricingPublicPlanCard{},
		CreditCards:       []HomePricingPublicCreditCard{},
	}
	for _, card := range cfg.SubscriptionCards {
		plan, ok := plansByID[card.SubscriptionPlanID]
		if !card.Enabled || !ok || !plan.ForSale {
			continue
		}
		out.SubscriptionCards = append(out.SubscriptionCards, HomePricingPublicPlanCard{
			ID:                 card.ID,
			Enabled:            card.Enabled,
			SortOrder:          card.SortOrder,
			SubscriptionPlanID: card.SubscriptionPlanID,
			Name:               card.Name,
			Description:        card.Description,
			Badge:              card.Badge,
			Period:             card.Period,
			Highlight:          card.Highlight,
			Metrics:            card.Metrics,
			Price:              plan.Price,
			OriginalPrice:      plan.OriginalPrice,
			ForSale:            plan.ForSale,
		})
	}
	for _, card := range cfg.CreditCards {
		if !card.Enabled {
			continue
		}
		creditedAmount := card.CreditedAmount
		if creditedAmount <= 0 {
			creditedAmount = calculateCreditedBalance(card.RechargeAmount, balanceRechargeMultiplier)
		}
		out.CreditCards = append(out.CreditCards, HomePricingPublicCreditCard{
			ID:             card.ID,
			Enabled:        card.Enabled,
			SortOrder:      card.SortOrder,
			RechargeAmount: card.RechargeAmount,
			CreditedAmount: creditedAmount,
			Name:           card.Name,
			Description:    card.Description,
			Badge:          card.Badge,
			Period:         card.Period,
			Highlight:      card.Highlight,
			Metrics:        card.Metrics,
			Price:          card.RechargeAmount,
		})
	}
	sort.SliceStable(out.SubscriptionCards, func(i, j int) bool {
		return out.SubscriptionCards[i].SortOrder < out.SubscriptionCards[j].SortOrder
	})
	sort.SliceStable(out.CreditCards, func(i, j int) bool {
		return out.CreditCards[i].SortOrder < out.CreditCards[j].SortOrder
	})
	return out
}

func (s *PaymentConfigService) homePricingBalanceRechargeMultiplier(ctx context.Context) float64 {
	cfg, err := s.GetPaymentConfig(ctx)
	if err != nil || cfg == nil {
		return defaultBalanceRechargeMultiplier
	}
	return normalizeBalanceRechargeMultiplier(cfg.BalanceRechargeMultiplier)
}

func (s *PaymentConfigService) homePricingPlansByID(ctx context.Context, cards []HomePricingSubscriptionCardConfig) map[int64]*dbent.SubscriptionPlan {
	if s == nil || s.entClient == nil {
		return map[int64]*dbent.SubscriptionPlan{}
	}
	ids := make([]int64, 0, len(cards))
	seen := make(map[int64]bool, len(cards))
	for _, card := range cards {
		if card.SubscriptionPlanID > 0 && !seen[card.SubscriptionPlanID] {
			seen[card.SubscriptionPlanID] = true
			ids = append(ids, card.SubscriptionPlanID)
		}
	}
	if len(ids) == 0 {
		return map[int64]*dbent.SubscriptionPlan{}
	}
	plans, err := s.entClient.SubscriptionPlan.Query().Where(subscriptionplan.IDIn(ids...)).All(ctx)
	if err != nil {
		return map[int64]*dbent.SubscriptionPlan{}
	}
	out := make(map[int64]*dbent.SubscriptionPlan, len(plans))
	for _, plan := range plans {
		out[int64(plan.ID)] = plan
	}
	return out
}

func (s *PaymentConfigService) validateHomePricingConfig(ctx context.Context, cfg *HomePricingConfig) error {
	if err := requireLocalizedText(cfg.Eyebrow, "eyebrow"); err != nil {
		return err
	}
	if err := requireLocalizedText(cfg.Title, "title"); err != nil {
		return err
	}
	if err := requireLocalizedText(cfg.Description, "description"); err != nil {
		return err
	}
	if err := requireLocalizedText(cfg.SubscriptionGroup.Title, "subscription_group.title"); err != nil {
		return err
	}
	if err := requireLocalizedText(cfg.CreditGroup.Title, "credit_group.title"); err != nil {
		return err
	}
	plansByID := s.homePricingPlansByID(ctx, cfg.SubscriptionCards)
	for i, card := range cfg.SubscriptionCards {
		prefix := fmt.Sprintf("subscription_cards[%d]", i)
		if card.SubscriptionPlanID <= 0 {
			return infraerrors.BadRequest("HOME_PRICING_PLAN_REQUIRED", prefix+" must bind a subscription plan")
		}
		if _, ok := plansByID[card.SubscriptionPlanID]; !ok {
			return infraerrors.BadRequest("HOME_PRICING_PLAN_NOT_FOUND", prefix+" references a missing subscription plan")
		}
		if err := requireLocalizedText(card.Name, prefix+".name"); err != nil {
			return err
		}
		if err := requireLocalizedText(card.Description, prefix+".description"); err != nil {
			return err
		}
	}
	for i, card := range cfg.CreditCards {
		prefix := fmt.Sprintf("credit_cards[%d]", i)
		if card.RechargeAmount <= 0 {
			return infraerrors.BadRequest("HOME_PRICING_RECHARGE_AMOUNT_INVALID", prefix+" recharge amount must be > 0")
		}
		if card.CreditedAmount <= 0 {
			return infraerrors.BadRequest("HOME_PRICING_CREDITED_AMOUNT_INVALID", prefix+" credited amount must be > 0")
		}
		if err := requireLocalizedText(card.Name, prefix+".name"); err != nil {
			return err
		}
		if err := requireLocalizedText(card.Description, prefix+".description"); err != nil {
			return err
		}
	}
	return nil
}

func requireLocalizedText(value HomePricingLocalizedText, field string) error {
	if strings.TrimSpace(value.Zh) == "" || strings.TrimSpace(value.En) == "" {
		return infraerrors.BadRequest("HOME_PRICING_I18N_REQUIRED", field+" requires both zh and en")
	}
	return nil
}
