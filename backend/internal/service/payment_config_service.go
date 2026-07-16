package service

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/paymentproviderinstance"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	SettingPaymentEnabled           = "payment_enabled"
	SettingMinRechargeAmount        = "MIN_RECHARGE_AMOUNT"
	SettingMaxRechargeAmount        = "MAX_RECHARGE_AMOUNT"
	SettingDailyRechargeLimit       = "DAILY_RECHARGE_LIMIT"
	SettingOrderTimeoutMinutes      = "ORDER_TIMEOUT_MINUTES"
	SettingMaxPendingOrders         = "MAX_PENDING_ORDERS"
	SettingEnabledPaymentTypes      = "ENABLED_PAYMENT_TYPES"
	SettingLoadBalanceStrategy      = "LOAD_BALANCE_STRATEGY"
	SettingBalancePayDisabled       = "BALANCE_PAYMENT_DISABLED"
	SettingBalanceRechargeMult      = "BALANCE_RECHARGE_MULTIPLIER"
	SettingRechargeFeeRate          = "RECHARGE_FEE_RATE"
	SettingProductNamePrefix        = "PRODUCT_NAME_PREFIX"
	SettingProductNameSuffix        = "PRODUCT_NAME_SUFFIX"
	SettingHelpImageURL             = "PAYMENT_HELP_IMAGE_URL"
	SettingHelpText                 = "PAYMENT_HELP_TEXT"
	SettingCancelRateLimitOn        = "CANCEL_RATE_LIMIT_ENABLED"
	SettingCancelRateLimitMax       = "CANCEL_RATE_LIMIT_MAX"
	SettingCancelWindowSize         = "CANCEL_RATE_LIMIT_WINDOW"
	SettingCancelWindowUnit         = "CANCEL_RATE_LIMIT_UNIT"
	SettingCancelWindowMode         = "CANCEL_RATE_LIMIT_WINDOW_MODE"
	SettingAlipayForceQRCode        = "ALIPAY_FORCE_QRCODE"
	SettingRenewalOfferEnabled      = "SUBSCRIPTION_RENEWAL_OFFER_ENABLED"
	SettingRenewalWindowDays        = "SUBSCRIPTION_RENEWAL_WINDOW_DAYS"
	SettingRenewalBeforeDays        = "SUBSCRIPTION_RENEWAL_BEFORE_DAYS"
	SettingRenewalAfterDays         = "SUBSCRIPTION_RENEWAL_AFTER_DAYS"
	SettingRenewalDiscountEnabled   = "SUBSCRIPTION_RENEWAL_DISCOUNT_ENABLED"
	SettingRenewalRolloverPercent   = "SUBSCRIPTION_RENEWAL_ROLLOVER_PERCENT"
	SettingRenewalDiscountPercent   = "SUBSCRIPTION_RENEWAL_DISCOUNT_PERCENT"
	SettingRenewalDiscountMinOrder  = "SUBSCRIPTION_RENEWAL_DISCOUNT_MIN_ORDER_AMOUNT"
	SettingRenewalDiscountMaxAmount = "SUBSCRIPTION_RENEWAL_DISCOUNT_MAX_AMOUNT"
	SettingRenewalRolloverEnabled   = "SUBSCRIPTION_RENEWAL_ROLLOVER_ENABLED"
	SettingRenewalRolloverMinUnused = "SUBSCRIPTION_RENEWAL_ROLLOVER_MIN_UNUSED_AMOUNT"
	SettingRenewalRolloverMaxAmount = "SUBSCRIPTION_RENEWAL_ROLLOVER_MAX_AMOUNT"
	SettingRenewalEmailEnabled      = "SUBSCRIPTION_RENEWAL_EMAIL_ENABLED"
	SettingRenewalEmailReminderDays = "SUBSCRIPTION_RENEWAL_EMAIL_REMINDER_DAYS"
)

// Default values for payment configuration settings.
const (
	defaultOrderTimeoutMin  = 30
	defaultMaxPendingOrders = 3
)

// PaymentConfig holds the payment system configuration.
type PaymentConfig struct {
	Enabled                   bool     `json:"enabled"`
	MinAmount                 float64  `json:"min_amount"`
	MaxAmount                 float64  `json:"max_amount"`
	DailyLimit                float64  `json:"daily_limit"`
	OrderTimeoutMin           int      `json:"order_timeout_minutes"`
	MaxPendingOrders          int      `json:"max_pending_orders"`
	EnabledTypes              []string `json:"enabled_payment_types"`
	BalanceDisabled           bool     `json:"balance_disabled"`
	BalanceRechargeMultiplier float64  `json:"balance_recharge_multiplier"`
	RechargeFeeRate           float64  `json:"recharge_fee_rate"`
	LoadBalanceStrategy       string   `json:"load_balance_strategy"`
	ProductNamePrefix         string   `json:"product_name_prefix"`
	ProductNameSuffix         string   `json:"product_name_suffix"`
	HelpImageURL              string   `json:"help_image_url"`
	HelpText                  string   `json:"help_text"`
	StripePublishableKey      string   `json:"stripe_publishable_key,omitempty"`

	// Cancel rate limit settings
	CancelRateLimitEnabled bool   `json:"cancel_rate_limit_enabled"`
	CancelRateLimitMax     int    `json:"cancel_rate_limit_max"`
	CancelRateLimitWindow  int    `json:"cancel_rate_limit_window"`
	CancelRateLimitUnit    string `json:"cancel_rate_limit_unit"`
	CancelRateLimitMode    string `json:"cancel_rate_limit_window_mode"`

	// Force Alipay mobile users to use QR code instead of mobile redirect
	AlipayForceQRCode bool `json:"alipay_force_qrcode"`

	RenewalOfferEnabled      bool    `json:"renewal_offer_enabled"`
	RenewalWindowDays        int     `json:"renewal_window_days"`
	RenewalBeforeDays        int     `json:"renewal_before_days"`
	RenewalAfterDays         int     `json:"renewal_after_days"`
	RenewalDiscountEnabled   bool    `json:"renewal_discount_enabled"`
	RenewalDiscountPercent   float64 `json:"renewal_discount_percent"`
	RenewalDiscountMinOrder  float64 `json:"renewal_discount_min_order_amount"`
	RenewalDiscountMaxAmount float64 `json:"renewal_discount_max_amount"`
	RenewalRolloverEnabled   bool    `json:"renewal_rollover_enabled"`
	RenewalRolloverPercent   float64 `json:"renewal_rollover_percent"`
	RenewalRolloverMinUnused float64 `json:"renewal_rollover_min_unused_amount"`
	RenewalRolloverMaxAmount float64 `json:"renewal_rollover_max_amount"`
	RenewalEmailEnabled      bool    `json:"renewal_email_enabled"`
	RenewalEmailReminderDays []int   `json:"renewal_email_reminder_days"`
}

// RenewalSettings is the dedicated admin-facing configuration for renewal campaigns.
// It intentionally remains separate from the general payment settings API.
type RenewalSettings struct {
	OfferEnabled      bool    `json:"offer_enabled"`
	BeforeExpiryDays  int     `json:"before_expiry_days"`
	AfterExpiryDays   int     `json:"after_expiry_days"`
	DiscountEnabled   bool    `json:"discount_enabled"`
	DiscountPercent   float64 `json:"discount_percent"`
	DiscountMinOrder  float64 `json:"discount_min_order_amount"`
	DiscountMaxAmount float64 `json:"discount_max_amount"`
	RolloverEnabled   bool    `json:"rollover_enabled"`
	RolloverPercent   float64 `json:"rollover_percent"`
	RolloverMinUnused float64 `json:"rollover_min_unused_amount"`
	RolloverMaxAmount float64 `json:"rollover_max_amount"`
	EmailEnabled      bool    `json:"email_enabled"`
	EmailReminderDays []int   `json:"email_reminder_days"`
}

// UpdatePaymentConfigRequest contains fields to update payment configuration.
type UpdatePaymentConfigRequest struct {
	Enabled                   *bool    `json:"enabled"`
	MinAmount                 *float64 `json:"min_amount"`
	MaxAmount                 *float64 `json:"max_amount"`
	DailyLimit                *float64 `json:"daily_limit"`
	OrderTimeoutMin           *int     `json:"order_timeout_minutes"`
	MaxPendingOrders          *int     `json:"max_pending_orders"`
	EnabledTypes              []string `json:"enabled_payment_types"`
	BalanceDisabled           *bool    `json:"balance_disabled"`
	BalanceRechargeMultiplier *float64 `json:"balance_recharge_multiplier"`
	RechargeFeeRate           *float64 `json:"recharge_fee_rate"`
	LoadBalanceStrategy       *string  `json:"load_balance_strategy"`
	ProductNamePrefix         *string  `json:"product_name_prefix"`
	ProductNameSuffix         *string  `json:"product_name_suffix"`
	HelpImageURL              *string  `json:"help_image_url"`
	HelpText                  *string  `json:"help_text"`

	// Cancel rate limit settings
	CancelRateLimitEnabled *bool   `json:"cancel_rate_limit_enabled"`
	CancelRateLimitMax     *int    `json:"cancel_rate_limit_max"`
	CancelRateLimitWindow  *int    `json:"cancel_rate_limit_window"`
	CancelRateLimitUnit    *string `json:"cancel_rate_limit_unit"`
	CancelRateLimitMode    *string `json:"cancel_rate_limit_window_mode"`

	// Force Alipay mobile users to use QR code instead of mobile redirect
	AlipayForceQRCode *bool `json:"alipay_force_qrcode"`

	RenewalOfferEnabled    *bool    `json:"renewal_offer_enabled"`
	RenewalWindowDays      *int     `json:"renewal_window_days"`
	RenewalRolloverPercent *float64 `json:"renewal_rollover_percent"`
	RenewalDiscountPercent *float64 `json:"renewal_discount_percent"`
	RenewalEmailEnabled    *bool    `json:"renewal_email_enabled"`

	VisibleMethodAlipaySource  *string `json:"payment_visible_method_alipay_source"`
	VisibleMethodWxpaySource   *string `json:"payment_visible_method_wxpay_source"`
	VisibleMethodAlipayEnabled *bool   `json:"payment_visible_method_alipay_enabled"`
	VisibleMethodWxpayEnabled  *bool   `json:"payment_visible_method_wxpay_enabled"`
}

// MethodLimits holds per-payment-type limits.
type MethodLimits struct {
	PaymentType string  `json:"payment_type"`
	DisplayName string  `json:"display_name,omitempty"`
	Currency    string  `json:"currency"`
	FeeRate     float64 `json:"fee_rate"`
	DailyLimit  float64 `json:"daily_limit"`
	DailyUsed   float64 `json:"daily_used"`
	DailyRemain float64 `json:"daily_remaining"`
	SingleMin   float64 `json:"single_min"`
	SingleMax   float64 `json:"single_max"`
	Available   bool    `json:"available"`
}

// MethodLimitsResponse is the full response for the user-facing /limits API.
// It includes per-method limits and the global widest range (union of all methods).
type MethodLimitsResponse struct {
	Methods   map[string]MethodLimits `json:"methods"`
	GlobalMin float64                 `json:"global_min"` // 0 = no minimum
	GlobalMax float64                 `json:"global_max"` // 0 = no maximum
}

type CreateProviderInstanceRequest struct {
	ProviderKey     string            `json:"provider_key"`
	Name            string            `json:"name"`
	Config          map[string]string `json:"config"`
	SupportedTypes  []string          `json:"supported_types"`
	Enabled         bool              `json:"enabled"`
	PaymentMode     string            `json:"payment_mode"`
	SortOrder       int               `json:"sort_order"`
	Limits          string            `json:"limits"`
	RefundEnabled   bool              `json:"refund_enabled"`
	AllowUserRefund bool              `json:"allow_user_refund"`
}

type UpdateProviderInstanceRequest struct {
	Name            *string           `json:"name"`
	Config          map[string]string `json:"config"`
	SupportedTypes  []string          `json:"supported_types"`
	Enabled         *bool             `json:"enabled"`
	PaymentMode     *string           `json:"payment_mode"`
	SortOrder       *int              `json:"sort_order"`
	Limits          *string           `json:"limits"`
	RefundEnabled   *bool             `json:"refund_enabled"`
	AllowUserRefund *bool             `json:"allow_user_refund"`
}
type CreatePlanRequest struct {
	GroupID       int64    `json:"group_id"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Price         float64  `json:"price"`
	OriginalPrice *float64 `json:"original_price"`
	ValidityDays  int      `json:"validity_days"`
	ValidityUnit  string   `json:"validity_unit"`
	Features      string   `json:"features"`
	ProductName   string   `json:"product_name"`
	ForSale       bool     `json:"for_sale"`
	SortOrder     int      `json:"sort_order"`
}

type UpdatePlanRequest struct {
	GroupID       *int64   `json:"group_id"`
	Name          *string  `json:"name"`
	Description   *string  `json:"description"`
	Price         *float64 `json:"price"`
	OriginalPrice *float64 `json:"original_price"`
	ValidityDays  *int     `json:"validity_days"`
	ValidityUnit  *string  `json:"validity_unit"`
	Features      *string  `json:"features"`
	ProductName   *string  `json:"product_name"`
	ForSale       *bool    `json:"for_sale"`
	SortOrder     *int     `json:"sort_order"`
}

// PaymentConfigService manages payment configuration and CRUD for
// provider instances, channels, and subscription plans.
type PaymentConfigService struct {
	entClient     *dbent.Client
	settingRepo   SettingRepository
	encryptionKey []byte
}

// NewPaymentConfigService creates a new PaymentConfigService.
func NewPaymentConfigService(entClient *dbent.Client, settingRepo SettingRepository, encryptionKey []byte) *PaymentConfigService {
	return &PaymentConfigService{entClient: entClient, settingRepo: settingRepo, encryptionKey: encryptionKey}
}

// IsPaymentEnabled returns whether the payment system is enabled.
func (s *PaymentConfigService) IsPaymentEnabled(ctx context.Context) bool {
	val, err := s.settingRepo.GetValue(ctx, SettingPaymentEnabled)
	if err != nil {
		return false
	}
	return val == "true"
}

// GetPaymentConfig returns the full payment configuration.
func (s *PaymentConfigService) GetPaymentConfig(ctx context.Context) (*PaymentConfig, error) {
	keys := []string{
		SettingPaymentEnabled, SettingMinRechargeAmount, SettingMaxRechargeAmount,
		SettingDailyRechargeLimit, SettingOrderTimeoutMinutes, SettingMaxPendingOrders,
		SettingEnabledPaymentTypes, SettingBalancePayDisabled, SettingBalanceRechargeMult, SettingRechargeFeeRate, SettingLoadBalanceStrategy,
		SettingProductNamePrefix, SettingProductNameSuffix,
		SettingHelpImageURL, SettingHelpText,
		SettingCancelRateLimitOn, SettingCancelRateLimitMax,
		SettingCancelWindowSize, SettingCancelWindowUnit, SettingCancelWindowMode,
		SettingAlipayForceQRCode,
		SettingRenewalOfferEnabled, SettingRenewalWindowDays, SettingRenewalBeforeDays, SettingRenewalAfterDays,
		SettingRenewalDiscountEnabled, SettingRenewalDiscountPercent, SettingRenewalDiscountMinOrder, SettingRenewalDiscountMaxAmount,
		SettingRenewalRolloverEnabled, SettingRenewalRolloverPercent, SettingRenewalRolloverMinUnused, SettingRenewalRolloverMaxAmount,
		SettingRenewalEmailEnabled, SettingRenewalEmailReminderDays,
		SettingPaymentVisibleMethodAlipayEnabled, SettingPaymentVisibleMethodAlipaySource,
		SettingPaymentVisibleMethodWxpayEnabled, SettingPaymentVisibleMethodWxpaySource,
	}
	vals, err := s.settingRepo.GetMultiple(ctx, keys)
	if err != nil {
		return nil, fmt.Errorf("get payment config settings: %w", err)
	}
	cfg := s.parsePaymentConfig(vals)
	// Load Stripe publishable key from the first enabled Stripe provider instance
	cfg.StripePublishableKey = s.getStripePublishableKey(ctx)
	return cfg, nil
}

func (s *PaymentConfigService) parsePaymentConfig(vals map[string]string) *PaymentConfig {
	legacyRenewalWindow := pcParseInt(vals[SettingRenewalWindowDays], 14)
	renewalBeforeDays := pcParseInt(vals[SettingRenewalBeforeDays], legacyRenewalWindow)
	renewalAfterDays := pcParseInt(vals[SettingRenewalAfterDays], legacyRenewalWindow)
	cfg := &PaymentConfig{
		Enabled:                   vals[SettingPaymentEnabled] == "true",
		MinAmount:                 pcParseFloat(vals[SettingMinRechargeAmount], 1),
		MaxAmount:                 pcParseFloat(vals[SettingMaxRechargeAmount], 0),
		DailyLimit:                pcParseFloat(vals[SettingDailyRechargeLimit], 0),
		OrderTimeoutMin:           pcParseInt(vals[SettingOrderTimeoutMinutes], defaultOrderTimeoutMin),
		MaxPendingOrders:          pcParseInt(vals[SettingMaxPendingOrders], defaultMaxPendingOrders),
		BalanceDisabled:           vals[SettingBalancePayDisabled] == "true",
		BalanceRechargeMultiplier: normalizeBalanceRechargeMultiplier(pcParseFloat(vals[SettingBalanceRechargeMult], defaultBalanceRechargeMultiplier)),
		RechargeFeeRate:           pcParseFloat(vals[SettingRechargeFeeRate], 0),
		LoadBalanceStrategy:       vals[SettingLoadBalanceStrategy],
		ProductNamePrefix:         vals[SettingProductNamePrefix],
		ProductNameSuffix:         vals[SettingProductNameSuffix],
		HelpImageURL:              vals[SettingHelpImageURL],
		HelpText:                  vals[SettingHelpText],

		CancelRateLimitEnabled: vals[SettingCancelRateLimitOn] == "true",
		CancelRateLimitMax:     pcParseInt(vals[SettingCancelRateLimitMax], 10),
		CancelRateLimitWindow:  pcParseInt(vals[SettingCancelWindowSize], 1),
		CancelRateLimitUnit:    vals[SettingCancelWindowUnit],
		CancelRateLimitMode:    vals[SettingCancelWindowMode],

		AlipayForceQRCode: vals[SettingAlipayForceQRCode] == "true",

		RenewalOfferEnabled:      vals[SettingRenewalOfferEnabled] == "true",
		RenewalWindowDays:        legacyRenewalWindow,
		RenewalBeforeDays:        renewalBeforeDays,
		RenewalAfterDays:         renewalAfterDays,
		RenewalDiscountEnabled:   vals[SettingRenewalDiscountEnabled] != "false",
		RenewalDiscountPercent:   pcParseFloat(vals[SettingRenewalDiscountPercent], 10),
		RenewalDiscountMinOrder:  pcParseFloat(vals[SettingRenewalDiscountMinOrder], 0),
		RenewalDiscountMaxAmount: pcParseFloat(vals[SettingRenewalDiscountMaxAmount], 0),
		RenewalRolloverEnabled:   vals[SettingRenewalRolloverEnabled] != "false",
		RenewalRolloverPercent:   pcParseFloat(vals[SettingRenewalRolloverPercent], 20),
		RenewalRolloverMinUnused: pcParseFloat(vals[SettingRenewalRolloverMinUnused], 0),
		RenewalRolloverMaxAmount: pcParseFloat(vals[SettingRenewalRolloverMaxAmount], 0),
		RenewalEmailEnabled:      vals[SettingRenewalEmailEnabled] != "false",
		RenewalEmailReminderDays: parseRenewalReminderDays(vals[SettingRenewalEmailReminderDays], renewalBeforeDays),
	}
	if cfg.LoadBalanceStrategy == "" {
		cfg.LoadBalanceStrategy = payment.DefaultLoadBalanceStrategy
	}
	if raw := vals[SettingEnabledPaymentTypes]; raw != "" {
		types := make([]string, 0, len(strings.Split(raw, ",")))
		for _, t := range strings.Split(raw, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				types = append(types, t)
			}
		}
		cfg.EnabledTypes = NormalizeVisibleMethods(types)
	}
	return cfg
}

// GetRenewalSettings returns the complete configuration owned by the renewal module.
func (s *PaymentConfigService) GetRenewalSettings(ctx context.Context) (*RenewalSettings, error) {
	cfg, err := s.GetPaymentConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &RenewalSettings{
		OfferEnabled:      cfg.RenewalOfferEnabled,
		BeforeExpiryDays:  cfg.RenewalBeforeDays,
		AfterExpiryDays:   cfg.RenewalAfterDays,
		DiscountEnabled:   cfg.RenewalDiscountEnabled,
		DiscountPercent:   cfg.RenewalDiscountPercent,
		DiscountMinOrder:  cfg.RenewalDiscountMinOrder,
		DiscountMaxAmount: cfg.RenewalDiscountMaxAmount,
		RolloverEnabled:   cfg.RenewalRolloverEnabled,
		RolloverPercent:   cfg.RenewalRolloverPercent,
		RolloverMinUnused: cfg.RenewalRolloverMinUnused,
		RolloverMaxAmount: cfg.RenewalRolloverMaxAmount,
		EmailEnabled:      cfg.RenewalEmailEnabled,
		EmailReminderDays: append([]int(nil), cfg.RenewalEmailReminderDays...),
	}, nil
}

// UpdateRenewalSettings validates and persists only renewal campaign settings.
func (s *PaymentConfigService) UpdateRenewalSettings(ctx context.Context, settings RenewalSettings) (*RenewalSettings, error) {
	if err := validateRenewalSettings(settings); err != nil {
		return nil, err
	}
	settings.EmailReminderDays = normalizeRenewalReminderDays(settings.EmailReminderDays)
	legacyWindow := settings.BeforeExpiryDays
	if settings.AfterExpiryDays > legacyWindow {
		legacyWindow = settings.AfterExpiryDays
	}
	values := map[string]string{
		SettingRenewalOfferEnabled:      strconv.FormatBool(settings.OfferEnabled),
		SettingRenewalWindowDays:        strconv.Itoa(legacyWindow),
		SettingRenewalBeforeDays:        strconv.Itoa(settings.BeforeExpiryDays),
		SettingRenewalAfterDays:         strconv.Itoa(settings.AfterExpiryDays),
		SettingRenewalDiscountEnabled:   strconv.FormatBool(settings.DiscountEnabled),
		SettingRenewalDiscountPercent:   strconv.FormatFloat(settings.DiscountPercent, 'f', 2, 64),
		SettingRenewalDiscountMinOrder:  strconv.FormatFloat(settings.DiscountMinOrder, 'f', 2, 64),
		SettingRenewalDiscountMaxAmount: strconv.FormatFloat(settings.DiscountMaxAmount, 'f', 2, 64),
		SettingRenewalRolloverEnabled:   strconv.FormatBool(settings.RolloverEnabled),
		SettingRenewalRolloverPercent:   strconv.FormatFloat(settings.RolloverPercent, 'f', 2, 64),
		SettingRenewalRolloverMinUnused: strconv.FormatFloat(settings.RolloverMinUnused, 'f', 10, 64),
		SettingRenewalRolloverMaxAmount: strconv.FormatFloat(settings.RolloverMaxAmount, 'f', 10, 64),
		SettingRenewalEmailEnabled:      strconv.FormatBool(settings.EmailEnabled),
		SettingRenewalEmailReminderDays: formatRenewalReminderDays(settings.EmailReminderDays),
	}
	if err := s.settingRepo.SetMultiple(ctx, values); err != nil {
		return nil, fmt.Errorf("update renewal settings: %w", err)
	}
	return s.GetRenewalSettings(ctx)
}

// getStripePublishableKey finds the publishable key from the first enabled Stripe provider instance.
func (s *PaymentConfigService) getStripePublishableKey(ctx context.Context) string {
	if s.entClient == nil {
		return ""
	}
	instances, err := s.entClient.PaymentProviderInstance.Query().
		Where(
			paymentproviderinstance.EnabledEQ(true),
			paymentproviderinstance.ProviderKeyEQ(payment.TypeStripe),
		).Limit(1).All(ctx)
	if err != nil || len(instances) == 0 {
		return ""
	}
	cfg, err := s.decryptConfig(instances[0].Config)
	if err != nil || cfg == nil {
		return ""
	}
	return cfg[payment.ConfigKeyPublishableKey]
}

// UpdatePaymentConfig updates the payment configuration settings.
// NOTE: This function exceeds 30 lines because each field requires an independent
// nil-check before serialisation — this is inherent to patch-style update patterns
// and cannot be meaningfully decomposed without introducing unnecessary abstraction.
func (s *PaymentConfigService) UpdatePaymentConfig(ctx context.Context, req UpdatePaymentConfigRequest) error {
	if req.BalanceRechargeMultiplier != nil {
		if math.IsNaN(*req.BalanceRechargeMultiplier) || math.IsInf(*req.BalanceRechargeMultiplier, 0) || *req.BalanceRechargeMultiplier <= 0 {
			return infraerrors.BadRequest("INVALID_BALANCE_RECHARGE_MULTIPLIER", "balance recharge multiplier must be greater than 0")
		}
	}
	if req.RechargeFeeRate != nil {
		v := *req.RechargeFeeRate
		if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || v > 100 {
			return infraerrors.BadRequest("INVALID_RECHARGE_FEE_RATE", "recharge fee rate must be between 0 and 100")
		}
		// Enforce max 2 decimal places
		if math.Round(v*100) != v*100 {
			return infraerrors.BadRequest("INVALID_RECHARGE_FEE_RATE", "recharge fee rate allows at most 2 decimal places")
		}
	}
	if req.RenewalWindowDays != nil && (*req.RenewalWindowDays < 1 || *req.RenewalWindowDays > 365) {
		return infraerrors.BadRequest("INVALID_RENEWAL_WINDOW_DAYS", "renewal window days must be between 1 and 365")
	}
	if value := req.RenewalRolloverPercent; value != nil && (math.IsNaN(*value) || math.IsInf(*value, 0) || *value < 0 || *value > 100) {
		return infraerrors.BadRequest("INVALID_RENEWAL_ROLLOVER_PERCENT", "renewal rollover percent must be between 0 and 100")
	}
	if value := req.RenewalDiscountPercent; value != nil && (math.IsNaN(*value) || math.IsInf(*value, 0) || *value < 0 || *value >= 100) {
		return infraerrors.BadRequest("INVALID_RENEWAL_DISCOUNT_PERCENT", "renewal discount percent must be at least 0 and less than 100")
	}
	m := map[string]string{
		SettingPaymentEnabled:                    formatBoolOrEmpty(req.Enabled),
		SettingMinRechargeAmount:                 formatPositiveFloat(req.MinAmount),
		SettingMaxRechargeAmount:                 formatPositiveFloat(req.MaxAmount),
		SettingDailyRechargeLimit:                formatPositiveFloat(req.DailyLimit),
		SettingOrderTimeoutMinutes:               formatPositiveInt(req.OrderTimeoutMin),
		SettingMaxPendingOrders:                  formatPositiveInt(req.MaxPendingOrders),
		SettingBalancePayDisabled:                formatBoolOrEmpty(req.BalanceDisabled),
		SettingBalanceRechargeMult:               formatPositiveFloat(req.BalanceRechargeMultiplier),
		SettingRechargeFeeRate:                   formatNonNegativeFloat(req.RechargeFeeRate),
		SettingLoadBalanceStrategy:               derefStr(req.LoadBalanceStrategy),
		SettingProductNamePrefix:                 derefStr(req.ProductNamePrefix),
		SettingProductNameSuffix:                 derefStr(req.ProductNameSuffix),
		SettingHelpImageURL:                      derefStr(req.HelpImageURL),
		SettingHelpText:                          derefStr(req.HelpText),
		SettingCancelRateLimitOn:                 formatBoolOrEmpty(req.CancelRateLimitEnabled),
		SettingCancelRateLimitMax:                formatPositiveInt(req.CancelRateLimitMax),
		SettingCancelWindowSize:                  formatPositiveInt(req.CancelRateLimitWindow),
		SettingCancelWindowUnit:                  derefStr(req.CancelRateLimitUnit),
		SettingCancelWindowMode:                  derefStr(req.CancelRateLimitMode),
		SettingAlipayForceQRCode:                 formatBoolOrEmpty(req.AlipayForceQRCode),
		SettingRenewalOfferEnabled:               formatBoolOrEmpty(req.RenewalOfferEnabled),
		SettingRenewalWindowDays:                 formatPositiveInt(req.RenewalWindowDays),
		SettingRenewalRolloverPercent:            formatNonNegativeFloat(req.RenewalRolloverPercent),
		SettingRenewalDiscountPercent:            formatNonNegativeFloat(req.RenewalDiscountPercent),
		SettingRenewalEmailEnabled:               formatBoolOrEmpty(req.RenewalEmailEnabled),
		SettingPaymentVisibleMethodAlipaySource:  derefStr(req.VisibleMethodAlipaySource),
		SettingPaymentVisibleMethodWxpaySource:   derefStr(req.VisibleMethodWxpaySource),
		SettingPaymentVisibleMethodAlipayEnabled: formatBoolOrEmpty(req.VisibleMethodAlipayEnabled),
		SettingPaymentVisibleMethodWxpayEnabled:  formatBoolOrEmpty(req.VisibleMethodWxpayEnabled),
	}
	if req.EnabledTypes != nil {
		m[SettingEnabledPaymentTypes] = strings.Join(req.EnabledTypes, ",")
	} else {
		m[SettingEnabledPaymentTypes] = ""
	}
	return s.settingRepo.SetMultiple(ctx, m)
}

func formatBoolOrEmpty(v *bool) string {
	if v == nil {
		return ""
	}
	return strconv.FormatBool(*v)
}

func validateRenewalSettings(settings RenewalSettings) error {
	if settings.BeforeExpiryDays < 0 || settings.BeforeExpiryDays > 365 || settings.AfterExpiryDays < 0 || settings.AfterExpiryDays > 365 {
		return infraerrors.BadRequest("INVALID_RENEWAL_WINDOW_DAYS", "renewal window days must be between 0 and 365")
	}
	if !validFiniteRange(settings.DiscountPercent, 0, 100, false) {
		return infraerrors.BadRequest("INVALID_RENEWAL_DISCOUNT_PERCENT", "renewal discount percent must be at least 0 and less than 100")
	}
	if !validFiniteRange(settings.RolloverPercent, 0, 100, true) {
		return infraerrors.BadRequest("INVALID_RENEWAL_ROLLOVER_PERCENT", "renewal rollover percent must be between 0 and 100")
	}
	for _, item := range []struct {
		value float64
		code  string
		name  string
	}{
		{settings.DiscountMinOrder, "INVALID_RENEWAL_DISCOUNT_MIN_ORDER", "discount minimum order amount"},
		{settings.DiscountMaxAmount, "INVALID_RENEWAL_DISCOUNT_MAX_AMOUNT", "discount maximum amount"},
		{settings.RolloverMinUnused, "INVALID_RENEWAL_ROLLOVER_MIN_UNUSED", "rollover minimum unused amount"},
		{settings.RolloverMaxAmount, "INVALID_RENEWAL_ROLLOVER_MAX_AMOUNT", "rollover maximum amount"},
	} {
		if math.IsNaN(item.value) || math.IsInf(item.value, 0) || item.value < 0 {
			return infraerrors.BadRequest(item.code, item.name+" must be a non-negative finite number")
		}
	}
	if len(settings.EmailReminderDays) > 20 {
		return infraerrors.BadRequest("INVALID_RENEWAL_EMAIL_REMINDER_DAYS", "at most 20 renewal email reminder days are allowed")
	}
	if settings.EmailEnabled && len(settings.EmailReminderDays) == 0 {
		return infraerrors.BadRequest("INVALID_RENEWAL_EMAIL_REMINDER_DAYS", "at least one renewal email reminder day is required when email is enabled")
	}
	for _, day := range settings.EmailReminderDays {
		if day < 0 || day > settings.BeforeExpiryDays {
			return infraerrors.BadRequest("INVALID_RENEWAL_EMAIL_REMINDER_DAYS", "renewal email reminder days must be within the pre-expiry renewal window")
		}
	}
	return nil
}

func validFiniteRange(value, minValue, maxValue float64, inclusiveMax bool) bool {
	if math.IsNaN(value) || math.IsInf(value, 0) || value < minValue {
		return false
	}
	if inclusiveMax {
		return value <= maxValue
	}
	return value < maxValue
}

func parseRenewalReminderDays(raw string, fallback int) []int {
	if strings.TrimSpace(raw) == "" {
		return []int{fallback}
	}
	days := make([]int, 0, 4)
	for _, item := range strings.Split(raw, ",") {
		day, err := strconv.Atoi(strings.TrimSpace(item))
		if err == nil && day >= 0 && day <= 365 {
			days = append(days, day)
		}
	}
	if len(days) == 0 {
		return []int{fallback}
	}
	return normalizeRenewalReminderDays(days)
}

func normalizeRenewalReminderDays(days []int) []int {
	seen := make(map[int]struct{}, len(days))
	result := make([]int, 0, len(days))
	for _, day := range days {
		if _, ok := seen[day]; ok {
			continue
		}
		seen[day] = struct{}{}
		result = append(result, day)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(result)))
	return result
}

func formatRenewalReminderDays(days []int) string {
	normalized := normalizeRenewalReminderDays(days)
	parts := make([]string, 0, len(normalized))
	for _, day := range normalized {
		parts = append(parts, strconv.Itoa(day))
	}
	return strings.Join(parts, ",")
}

func formatPositiveFloat(v *float64) string {
	if v == nil || *v <= 0 {
		return "" // empty → parsePaymentConfig uses default
	}
	return strconv.FormatFloat(*v, 'f', 2, 64)
}

// formatPositiveFloatExact 保留完整精度，用于汇率等对小数位敏感的配置。
func formatPositiveFloatExact(v *float64) string {
	if v == nil || *v <= 0 {
		return "" // empty → parsePaymentConfig 视为未配置（换算关闭）
	}
	return strconv.FormatFloat(*v, 'f', -1, 64)
}

func formatNonNegativeFloat(v *float64) string {
	if v == nil || *v < 0 {
		return ""
	}
	return strconv.FormatFloat(*v, 'f', 2, 64)
}

func formatPositiveInt(v *int) string {
	if v == nil || *v <= 0 {
		return ""
	}
	return strconv.Itoa(*v)
}

func derefStr(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func splitTypes(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func joinTypes(types []string) string {
	return strings.Join(types, ",")
}

func pcParseFloat(s string, defaultVal float64) float64 {
	if s == "" {
		return defaultVal
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return defaultVal
	}
	return v
}

func pcParseInt(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return v
}

func buildVisibleMethodSourceAvailability(instances []*dbent.PaymentProviderInstance) map[string]bool {
	available := make(map[string]bool, 4)
	for _, inst := range instances {
		switch inst.ProviderKey {
		case payment.TypeAlipay:
			if inst.SupportedTypes == "" || payment.InstanceSupportsType(inst.SupportedTypes, payment.TypeAlipay) || payment.InstanceSupportsType(inst.SupportedTypes, payment.TypeAlipayDirect) {
				available[VisibleMethodSourceOfficialAlipay] = true
			}
		case payment.TypeWxpay:
			if inst.SupportedTypes == "" || payment.InstanceSupportsType(inst.SupportedTypes, payment.TypeWxpay) || payment.InstanceSupportsType(inst.SupportedTypes, payment.TypeWxpayDirect) {
				available[VisibleMethodSourceOfficialWechat] = true
			}
		case payment.TypeEasyPay:
			for _, supportedType := range splitTypes(inst.SupportedTypes) {
				switch NormalizeVisibleMethod(supportedType) {
				case payment.TypeAlipay:
					available[VisibleMethodSourceEasyPayAlipay] = true
				case payment.TypeWxpay:
					available[VisibleMethodSourceEasyPayWechat] = true
				}
			}
		}
	}
	return available
}

func applyVisibleMethodRoutingToEnabledTypes(base []string, vals map[string]string, available map[string]bool) []string {
	shouldExpose := map[string]bool{
		payment.TypeAlipay: visibleMethodShouldBeExposed(payment.TypeAlipay, vals, available),
		payment.TypeWxpay:  visibleMethodShouldBeExposed(payment.TypeWxpay, vals, available),
	}

	seen := make(map[string]struct{}, len(base)+2)
	out := make([]string, 0, len(base)+2)
	appendType := func(paymentType string) {
		paymentType = NormalizeVisibleMethod(paymentType)
		if paymentType == "" {
			return
		}
		if _, ok := seen[paymentType]; ok {
			return
		}
		seen[paymentType] = struct{}{}
		out = append(out, paymentType)
	}

	for _, paymentType := range base {
		visibleMethod := NormalizeVisibleMethod(paymentType)
		switch visibleMethod {
		case payment.TypeAlipay, payment.TypeWxpay:
			if shouldExpose[visibleMethod] {
				appendType(visibleMethod)
			}
		default:
			appendType(visibleMethod)
		}
	}

	for _, visibleMethod := range []string{payment.TypeAlipay, payment.TypeWxpay} {
		if shouldExpose[visibleMethod] {
			appendType(visibleMethod)
		}
	}
	return out
}

func visibleMethodShouldBeExposed(method string, vals map[string]string, available map[string]bool) bool {
	enabledKey := visibleMethodEnabledSettingKey(method)
	sourceKey := visibleMethodSourceSettingKey(method)
	if enabledKey == "" || sourceKey == "" || vals[enabledKey] != "true" {
		return false
	}
	source := NormalizeVisibleMethodSource(method, vals[sourceKey])
	return source != "" && available[source]
}
