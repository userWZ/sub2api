package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

type imageWorkbenchRequestContextKey struct{}

const (
	ContextKeyImageWorkbenchRequest  = "image_workbench_request"
	SettingKeyImageWorkbenchAccounts = "image_workbench_accounts"
)

type ImageWorkbenchAccountSettings struct {
	Enabled    bool    `json:"enabled"`
	AccountIDs []int64 `json:"account_ids"`
}

func WithImageWorkbenchRequestContext(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, imageWorkbenchRequestContextKey{}, true)
}

func IsImageWorkbenchRequestContext(ctx context.Context) bool {
	value, _ := ctx.Value(imageWorkbenchRequestContextKey{}).(bool)
	return value
}

func DefaultImageWorkbenchAccountSettings() *ImageWorkbenchAccountSettings {
	return &ImageWorkbenchAccountSettings{
		Enabled:    false,
		AccountIDs: []int64{},
	}
}

func (s *SettingService) GetImageWorkbenchAccountSettings(ctx context.Context) (*ImageWorkbenchAccountSettings, error) {
	value, err := s.settingRepo.GetValue(ctx, SettingKeyImageWorkbenchAccounts)
	if err != nil {
		if errors.Is(err, ErrSettingNotFound) {
			return DefaultImageWorkbenchAccountSettings(), nil
		}
		return nil, fmt.Errorf("get image workbench account settings: %w", err)
	}
	if value == "" {
		return DefaultImageWorkbenchAccountSettings(), nil
	}

	var settings ImageWorkbenchAccountSettings
	if err := json.Unmarshal([]byte(value), &settings); err != nil {
		return DefaultImageWorkbenchAccountSettings(), nil
	}
	settings.AccountIDs = normalizeImageWorkbenchAccountIDs(settings.AccountIDs)
	if len(settings.AccountIDs) == 0 {
		settings.Enabled = false
	}
	return &settings, nil
}

func (s *SettingService) SetImageWorkbenchAccountSettings(ctx context.Context, settings *ImageWorkbenchAccountSettings) error {
	if settings == nil {
		return fmt.Errorf("settings cannot be nil")
	}
	normalized := &ImageWorkbenchAccountSettings{
		Enabled:    settings.Enabled,
		AccountIDs: normalizeImageWorkbenchAccountIDs(settings.AccountIDs),
	}
	if len(normalized.AccountIDs) == 0 {
		normalized.Enabled = false
	}
	data, err := json.Marshal(normalized)
	if err != nil {
		return fmt.Errorf("marshal image workbench account settings: %w", err)
	}
	return s.settingRepo.Set(ctx, SettingKeyImageWorkbenchAccounts, string(data))
}

func normalizeImageWorkbenchAccountIDs(accountIDs []int64) []int64 {
	if len(accountIDs) == 0 {
		return []int64{}
	}
	seen := make(map[int64]struct{}, len(accountIDs))
	normalized := make([]int64, 0, len(accountIDs))
	for _, id := range accountIDs {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		normalized = append(normalized, id)
	}
	return normalized
}

func imageWorkbenchAccountAllowed(allowlist map[int64]struct{}, accountID int64) bool {
	if len(allowlist) == 0 {
		return true
	}
	_, ok := allowlist[accountID]
	return ok
}

func (s *OpenAIGatewayService) imageWorkbenchAccountAllowlist(ctx context.Context) map[int64]struct{} {
	if !IsImageWorkbenchRequestContext(ctx) || s == nil || s.settingService == nil {
		return nil
	}
	settings, err := s.settingService.GetImageWorkbenchAccountSettings(ctx)
	if err != nil || settings == nil || !settings.Enabled || len(settings.AccountIDs) == 0 {
		return nil
	}
	allowlist := make(map[int64]struct{}, len(settings.AccountIDs))
	for _, id := range settings.AccountIDs {
		if id > 0 {
			allowlist[id] = struct{}{}
		}
	}
	if len(allowlist) == 0 {
		return nil
	}
	return allowlist
}
