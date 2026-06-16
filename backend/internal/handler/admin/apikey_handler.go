package admin

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

const (
	managedKeyMarker        = "[managed-key]"
	managedKeySearchTerm    = "managed-key"
	managedKeyEmailDomain   = "managed.local"
	managedPremiumGroupName = "codex-managed-premium"
	defaultManagedKeyFunds  = 1000
	defaultManagedKeyDays   = 30
)

// AdminAPIKeyHandler handles admin API key management
type AdminAPIKeyHandler struct {
	adminService  service.AdminService
	apiKeyService *service.APIKeyService
}

// NewAdminAPIKeyHandler creates a new admin API key handler
func NewAdminAPIKeyHandler(adminService service.AdminService, apiKeyService *service.APIKeyService) *AdminAPIKeyHandler {
	return &AdminAPIKeyHandler{
		adminService:  adminService,
		apiKeyService: apiKeyService,
	}
}

// AdminUpdateAPIKeyGroupRequest represents the request to update an API key.
type AdminUpdateAPIKeyGroupRequest struct {
	GroupID             *int64 `json:"group_id"`               // nil=不修改, 0=解绑, >0=绑定到目标分组
	ResetRateLimitUsage *bool  `json:"reset_rate_limit_usage"` // true=重置 5h/1d/7d 限速用量
}

// CreateManagedKeyRequest creates an internal managed user plus a customer-facing API key.
type CreateManagedKeyRequest struct {
	CustomerName  string   `json:"customer_name" binding:"required"`
	Contact       string   `json:"contact"`
	KeyName       string   `json:"key_name"`
	GroupID       *int64   `json:"group_id"`
	Balance       *float64 `json:"balance"`
	Concurrency   int      `json:"concurrency"`
	RPMLimit      int      `json:"rpm_limit"`
	Quota         float64  `json:"quota"`
	ExpiresInDays *int     `json:"expires_in_days"`
	CustomKey     *string  `json:"custom_key"`
	IPWhitelist   []string `json:"ip_whitelist"`
	IPBlacklist   []string `json:"ip_blacklist"`
	RateLimit5h   float64  `json:"rate_limit_5h"`
	RateLimit1d   float64  `json:"rate_limit_1d"`
	RateLimit7d   float64  `json:"rate_limit_7d"`
	Notes         string   `json:"notes"`
}

type ManagedKey struct {
	User   *dto.AdminUser `json:"user"`
	APIKey *dto.APIKey    `json:"api_key"`
}

type ManagedKeyDelivery struct {
	APIKey              string `json:"api_key"`
	AuthorizationHeader string `json:"authorization_header"`
	BaseURL             string `json:"base_url"`
	OpenAIBaseURL       string `json:"openai_base_url"`
	ClaudeBaseURL       string `json:"claude_base_url"`
	GeminiBaseURL       string `json:"gemini_base_url"`
}

type ManagedKeyResponse struct {
	User     *dto.AdminUser     `json:"user"`
	APIKey   *dto.APIKey        `json:"api_key"`
	Delivery ManagedKeyDelivery `json:"delivery"`
}

// UpdateGroup handles updating an API key's admin-managed fields.
// PUT /api/v1/admin/api-keys/:id
func (h *AdminAPIKeyHandler) UpdateGroup(c *gin.Context) {
	keyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid API key ID")
		return
	}

	var req AdminUpdateAPIKeyGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	var resetKey *service.APIKey
	if req.ResetRateLimitUsage != nil && *req.ResetRateLimitUsage {
		resetKey, err = h.adminService.AdminResetAPIKeyRateLimitUsage(c.Request.Context(), keyID)
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
	}

	result, err := h.adminService.AdminUpdateAPIKeyGroupID(c.Request.Context(), keyID, req.GroupID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if resetKey != nil && req.GroupID == nil {
		result.APIKey = resetKey
	}

	resp := struct {
		APIKey                 *dto.APIKey `json:"api_key"`
		AutoGrantedGroupAccess bool        `json:"auto_granted_group_access"`
		GrantedGroupID         *int64      `json:"granted_group_id,omitempty"`
		GrantedGroupName       string      `json:"granted_group_name,omitempty"`
	}{
		APIKey:                 dto.APIKeyFromService(result.APIKey),
		AutoGrantedGroupAccess: result.AutoGrantedGroupAccess,
		GrantedGroupID:         result.GrantedGroupID,
		GrantedGroupName:       result.GrantedGroupName,
	}
	response.Success(c, resp)
}

// ListManagedKeys lists managed users with their primary API key.
// GET /api/v1/admin/managed-keys
func (h *AdminAPIKeyHandler) ListManagedKeys(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	users, total, err := h.adminService.ListUsers(
		c.Request.Context(),
		page,
		pageSize,
		service.UserListFilters{Role: service.RoleUser, CustomerType: service.CustomerTypeManaged},
		"created_at",
		"desc",
	)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	items := make([]ManagedKey, 0, len(users))
	for i := range users {
		user := users[i]
		keys, _, keyErr := h.adminService.GetUserAPIKeys(c.Request.Context(), user.ID, 1, 1, "created_at", "desc")
		if keyErr != nil {
			response.ErrorFrom(c, keyErr)
			return
		}
		var keyDTO *dto.APIKey
		if len(keys) > 0 {
			keyDTO = dto.APIKeyFromService(&keys[0])
		}
		items = append(items, ManagedKey{
			User:   dto.UserFromServiceAdmin(&user),
			APIKey: keyDTO,
		})
	}

	response.Paginated(c, items, total, page, pageSize)
}

// CreateManagedKey creates an internal managed user and a customer-facing API key.
// POST /api/v1/admin/managed-keys
func (h *AdminAPIKeyHandler) CreateManagedKey(c *gin.Context) {
	if h.apiKeyService == nil {
		response.InternalError(c, "API key service is unavailable")
		return
	}

	var req CreateManagedKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	normalized, err := normalizeManagedKeyRequest(req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	managedGroup, err := h.resolveManagedKeyGroup(c.Request.Context(), normalized.GroupID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	normalized.GroupID = &managedGroup.ID

	password, err := randomHex(24)
	if err != nil {
		response.InternalError(c, "failed to generate managed user password")
		return
	}
	suffix, err := randomHex(6)
	if err != nil {
		response.InternalError(c, "failed to generate managed user email")
		return
	}

	email := fmt.Sprintf("%s-%s@%s", managedKeySearchTerm, suffix, managedKeyEmailDomain)
	user, err := h.adminService.CreateUser(c.Request.Context(), &service.CreateUserInput{
		Email:         email,
		Password:      password,
		Username:      normalized.CustomerName,
		Notes:         buildManagedKeyNotes(normalized),
		CustomerType:  service.CustomerTypeManaged,
		Balance:       &normalized.balanceValue,
		Concurrency:   normalized.Concurrency,
		RPMLimit:      normalized.RPMLimit,
		AllowedGroups: []int64{managedGroup.ID},
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	apiKey, err := h.apiKeyService.Create(c.Request.Context(), user.ID, service.CreateAPIKeyRequest{
		Name:          normalized.KeyName,
		GroupID:       normalized.GroupID,
		CustomKey:     normalized.CustomKey,
		IPWhitelist:   normalized.IPWhitelist,
		IPBlacklist:   normalized.IPBlacklist,
		Quota:         normalized.Quota,
		ExpiresInDays: normalized.ExpiresInDays,
		RateLimit5h:   normalized.RateLimit5h,
		RateLimit1d:   normalized.RateLimit1d,
		RateLimit7d:   normalized.RateLimit7d,
	})
	if err != nil {
		_ = h.adminService.DeleteUser(c.Request.Context(), user.ID)
		response.ErrorFrom(c, err)
		return
	}

	response.Created(c, ManagedKeyResponse{
		User:     dto.UserFromServiceAdmin(user),
		APIKey:   dto.APIKeyFromService(apiKey),
		Delivery: buildManagedKeyDelivery(c, apiKey.Key),
	})
}

func (h *AdminAPIKeyHandler) resolveManagedKeyGroup(ctx context.Context, groupID *int64) (*service.Group, error) {
	if groupID != nil && *groupID > 0 {
		group, err := h.adminService.GetGroup(ctx, *groupID)
		if err != nil {
			return nil, err
		}
		if err := validateManagedKeyGroup(group); err != nil {
			return nil, err
		}
		return group, nil
	}

	groups, err := h.adminService.GetAllGroups(ctx)
	if err != nil {
		return nil, err
	}
	for i := range groups {
		group := groups[i]
		if strings.EqualFold(strings.TrimSpace(group.Name), managedPremiumGroupName) {
			if err := validateManagedKeyGroup(&group); err != nil {
				return nil, fmt.Errorf("%s is not a valid managed key group: %w", managedPremiumGroupName, err)
			}
			return &group, nil
		}
	}
	return nil, fmt.Errorf("managed key group %q is required; create an active standard group first", managedPremiumGroupName)
}

func validateManagedKeyGroup(group *service.Group) error {
	if group == nil || group.ID <= 0 {
		return fmt.Errorf("managed key group is required")
	}
	if !group.IsActive() {
		return fmt.Errorf("managed key group must be active")
	}
	if group.IsSubscriptionType() {
		return fmt.Errorf("managed key group must be a standard group, not a subscription group")
	}
	return nil
}

// GetManagedKeyDelivery returns customer-facing delivery details for an existing managed key.
// GET /api/v1/admin/managed-keys/:id/delivery
func (h *AdminAPIKeyHandler) GetManagedKeyDelivery(c *gin.Context) {
	user, apiKey, ok := h.loadManagedUserPrimaryKey(c)
	if !ok {
		return
	}
	response.Success(c, ManagedKeyResponse{
		User:     dto.UserFromServiceAdmin(user),
		APIKey:   dto.APIKeyFromService(apiKey),
		Delivery: buildManagedKeyDelivery(c, apiKey.Key),
	})
}

type normalizedManagedKeyRequest struct {
	CreateManagedKeyRequest
	balanceValue float64
}

func normalizeManagedKeyRequest(req CreateManagedKeyRequest) (*normalizedManagedKeyRequest, error) {
	req.CustomerName = strings.TrimSpace(req.CustomerName)
	req.Contact = strings.TrimSpace(req.Contact)
	req.KeyName = strings.TrimSpace(req.KeyName)
	req.Notes = strings.TrimSpace(req.Notes)
	if req.CustomerName == "" {
		return nil, fmt.Errorf("customer_name is required")
	}
	if req.KeyName == "" {
		req.KeyName = req.CustomerName + " API Key"
	}
	if req.GroupID != nil && *req.GroupID <= 0 {
		req.GroupID = nil
	}
	if req.Concurrency <= 0 {
		req.Concurrency = 1
	}
	if req.RPMLimit < 0 {
		return nil, fmt.Errorf("rpm_limit cannot be negative")
	}
	if req.Quota < 0 {
		return nil, fmt.Errorf("quota cannot be negative")
	}
	if req.RateLimit5h < 0 || req.RateLimit1d < 0 || req.RateLimit7d < 0 {
		return nil, fmt.Errorf("rate limits cannot be negative")
	}
	if req.ExpiresInDays != nil && *req.ExpiresInDays < 0 {
		return nil, fmt.Errorf("expires_in_days cannot be negative")
	}
	if req.ExpiresInDays == nil || *req.ExpiresInDays == 0 {
		defaultDays := defaultManagedKeyDays
		req.ExpiresInDays = &defaultDays
	}
	if req.CustomKey != nil {
		trimmed := strings.TrimSpace(*req.CustomKey)
		if trimmed == "" {
			req.CustomKey = nil
		} else {
			req.CustomKey = &trimmed
		}
	}

	balance := float64(defaultManagedKeyFunds)
	if req.Balance != nil {
		if *req.Balance < 0 {
			return nil, fmt.Errorf("balance cannot be negative")
		}
		balance = *req.Balance
	}

	return &normalizedManagedKeyRequest{
		CreateManagedKeyRequest: req,
		balanceValue:            balance,
	}, nil
}

func buildManagedKeyNotes(req *normalizedManagedKeyRequest) string {
	lines := []string{
		managedKeyMarker,
		"customer: " + compactNoteLine(req.CustomerName),
	}
	if req.Contact != "" {
		lines = append(lines, "contact: "+compactNoteLine(req.Contact))
	}
	if req.Notes != "" {
		lines = append(lines, "", req.Notes)
	}
	return strings.Join(lines, "\n")
}

func compactNoteLine(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func buildManagedKeyDelivery(c *gin.Context, key string) ManagedKeyDelivery {
	baseURL := requestBaseURL(c)
	v1Base := strings.TrimRight(baseURL, "/") + "/v1"
	return ManagedKeyDelivery{
		APIKey:              key,
		AuthorizationHeader: "Bearer " + key,
		BaseURL:             baseURL,
		OpenAIBaseURL:       v1Base,
		ClaudeBaseURL:       v1Base,
		GeminiBaseURL:       strings.TrimRight(baseURL, "/") + "/v1beta",
	}
}

func requestBaseURL(c *gin.Context) string {
	proto := strings.TrimSpace(c.GetHeader("X-Forwarded-Proto"))
	if proto == "" {
		proto = strings.TrimSpace(c.GetHeader("X-Forwarded-Scheme"))
	}
	if proto == "" {
		proto = "http"
		if c.Request != nil && c.Request.TLS != nil {
			proto = "https"
		}
	}

	host := strings.TrimSpace(c.GetHeader("X-Forwarded-Host"))
	if host == "" && c.Request != nil {
		host = c.Request.Host
	}
	if host == "" {
		return ""
	}
	return proto + "://" + host
}

func (h *AdminAPIKeyHandler) loadManagedUserPrimaryKey(c *gin.Context) (*service.User, *service.APIKey, bool) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid managed key ID")
		return nil, nil, false
	}
	user, err := h.adminService.GetUser(c.Request.Context(), userID)
	if err != nil {
		response.ErrorFrom(c, err)
		return nil, nil, false
	}
	if user.CustomerType != service.CustomerTypeManaged {
		response.NotFound(c, "Managed key not found")
		return nil, nil, false
	}
	keys, _, err := h.adminService.GetUserAPIKeys(c.Request.Context(), user.ID, 1, 1, "created_at", "desc")
	if err != nil {
		response.ErrorFrom(c, err)
		return nil, nil, false
	}
	if len(keys) == 0 {
		response.NotFound(c, "Managed key API key not found")
		return nil, nil, false
	}
	return user, &keys[0], true
}

func randomHex(bytesLen int) (string, error) {
	buf := make([]byte, bytesLen)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
