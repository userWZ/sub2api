package middleware

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ip"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// NewAPIKeyAuthMiddleware 创建 API Key 认证中间件
func NewAPIKeyAuthMiddleware(apiKeyService *service.APIKeyService, subscriptionService *service.SubscriptionService, cfg *config.Config) APIKeyAuthMiddleware {
	return APIKeyAuthMiddleware(apiKeyAuthWithSubscription(apiKeyService, subscriptionService, cfg))
}

// apiKeyAuthWithSubscription API Key认证中间件（支持订阅验证）
//
// 中间件职责分为两层：
//   - 鉴权（Authentication）：验证 Key 有效性、用户状态、IP 限制 —— 始终执行
//   - 计费执行（Billing Enforcement）：过期/配额/订阅/余额检查 —— skipBilling 时整块跳过
//
// /v1/usage 端点只需鉴权，不需要计费执行（允许过期/配额耗尽的 Key 查询自身用量）。
func apiKeyAuthWithSubscription(apiKeyService *service.APIKeyService, subscriptionService *service.SubscriptionService, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// ── 1. 提取 API Key ──────────────────────────────────────────

		queryKey := strings.TrimSpace(c.Query("key"))
		queryApiKey := strings.TrimSpace(c.Query("api_key"))
		if queryKey != "" || queryApiKey != "" {
			AbortWithError(c, 400, "api_key_in_query_deprecated", "API key in query parameter is deprecated. Please use Authorization header instead.")
			return
		}

		// 尝试从Authorization header中提取API key (Bearer scheme)
		authHeader := c.GetHeader("Authorization")
		var apiKeyString string

		if authHeader != "" {
			// 验证Bearer scheme
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
				apiKeyString = strings.TrimSpace(parts[1])
			}
		}

		// 如果Authorization header中没有，尝试从x-api-key header中提取
		if apiKeyString == "" {
			apiKeyString = c.GetHeader("x-api-key")
		}

		// 如果x-api-key header中没有，尝试从x-goog-api-key header中提取（Gemini CLI兼容）
		if apiKeyString == "" {
			apiKeyString = c.GetHeader("x-goog-api-key")
		}

		// 如果所有header都没有API key
		if apiKeyString == "" {
			AbortWithError(c, 401, "API_KEY_REQUIRED", "API key is required in Authorization header (Bearer scheme), x-api-key header, or x-goog-api-key header")
			return
		}

		// ── 2. 验证 Key 存在 ─────────────────────────────────────────

		apiKey, err := apiKeyService.GetByKey(c.Request.Context(), apiKeyString)
		if err != nil {
			if errors.Is(err, service.ErrAPIKeyNotFound) {
				AbortWithError(c, 401, "INVALID_API_KEY", "Invalid API key")
				return
			}
			AbortWithError(c, 500, "INTERNAL_ERROR", "Failed to validate API key")
			return
		}

		// apiKey 已加载（含 User/Group）。即便后续因分组停用/Key 停用/用户停用/
		// IP 限制等早退中断，也让 Ops 错误日志能回退取到 user/group/platform。
		SetOpsFallbackAPIKey(c, apiKey)

		// ── 3. 基础鉴权（始终执行） ─────────────────────────────────

		// disabled / 未知状态 → 无条件拦截（expired 和 quota_exhausted 留给计费阶段）
		if !apiKey.IsActive() &&
			apiKey.Status != service.StatusAPIKeyExpired &&
			apiKey.Status != service.StatusAPIKeyQuotaExhausted {
			AbortWithError(c, 401, "API_KEY_DISABLED", "API key is disabled")
			return
		}

		// 检查 IP 限制（白名单/黑名单）
		// 注意：错误信息故意模糊，避免暴露具体的 IP 限制机制
		if len(apiKey.IPWhitelist) > 0 || len(apiKey.IPBlacklist) > 0 {
			clientIP := ip.GetTrustedClientIP(c)
			if cfg.TrustForwardedIPForAPIKeyACL() {
				clientIP = ip.GetClientIP(c)
			}
			allowed, _ := ip.CheckIPRestrictionWithCompiledRules(clientIP, apiKey.CompiledIPWhitelist, apiKey.CompiledIPBlacklist)
			if !allowed {
				if clientIP == "" {
					clientIP = "unknown"
				}
				service.MarkOpsClientBusinessLimited(c, service.OpsClientBusinessLimitedReasonIPRestriction)
				AbortWithError(c, 403, "ACCESS_DENIED", fmt.Sprintf("Access denied. Your IP is %s", clientIP))
				return
			}
		}

		// 检查关联的用户
		if apiKey.User == nil {
			AbortWithError(c, 401, "USER_NOT_FOUND", "User associated with API key not found")
			return
		}

		// 检查用户状态
		if !apiKey.User.IsActive() {
			AbortWithError(c, 401, "USER_INACTIVE", "User account is not active")
			return
		}
		if abortIfAPIKeyGroupUnavailable(c, apiKey) {
			return
		}
		if abortIfAPIKeyGroupNotAllowed(c, apiKey) {
			return
		}
		ctx := context.WithValue(c.Request.Context(), ctxkey.UserID, apiKey.User.ID)
		c.Request = c.Request.WithContext(ctx)

		// ── 4. SimpleMode → early return ─────────────────────────────

		if cfg.RunMode == config.RunModeSimple {
			c.Set(string(ContextKeyAPIKey), apiKey)
			c.Set(string(ContextKeyUser), AuthSubject{
				UserID:      apiKey.User.ID,
				Concurrency: apiKey.User.Concurrency,
			})
			c.Set(string(ContextKeyUserRole), apiKey.User.Role)
			setGroupContext(c, apiKey.Group)
			_ = apiKeyService.TouchLastUsed(c.Request.Context(), apiKey.ID)
			c.Next()
			return
		}

		// ── 5. 加载订阅（订阅模式时始终加载） ───────────────────────

		// skipBilling: /v1/usage 只需鉴权，跳过所有计费执行
		skipBilling := c.Request.URL.Path == "/v1/usage"
		entitlementReq := requestEntitlementRequest(c)
		setEntitlementRequestContext(c, entitlementReq)

		var subscription *service.UserSubscription
		automaticSubscriptionResolved := false
		if apiKey.Group == nil || apiKey.GroupID == nil {
			resolvedSub, resolveErr := resolveUnboundAPIKeyEntitlement(c.Request.Context(), apiKeyService, subscriptionService, apiKey, entitlementReq)
			if resolveErr != nil {
				abortEntitlementResolutionError(c, apiKey, entitlementReq, resolveErr, "Failed to resolve API key entitlement")
				return
			}
			subscription = resolvedSub
			automaticSubscriptionResolved = subscription != nil
		}
		isSubscriptionType := apiKey.Group != nil && apiKey.Group.IsSubscriptionType()

		if isSubscriptionType && subscriptionService != nil {
			if subscription == nil {
				sub, subErr := subscriptionService.GetActiveSubscription(
					c.Request.Context(),
					apiKey.User.ID,
					apiKey.Group.ID,
				)
				if subErr != nil {
					if !skipBilling {
						if ok, fallbackErr := resolveExplicitBalanceFallbackEntitlement(c.Request.Context(), apiKeyService, apiKey, entitlementReq); fallbackErr != nil {
							abortEntitlementResolutionError(c, apiKey, entitlementReq, fallbackErr, "Failed to resolve balance fallback entitlement")
							return
						} else if ok {
							isSubscriptionType = false
						} else {
							AbortWithError(c, 403, "SUBSCRIPTION_NOT_FOUND", "No active subscription found for this group")
							return
						}
					}
					// skipBilling: 订阅不存在也放行，handler 会返回可用的数据
				} else {
					subscription = sub
				}
			}
		}

		// ── 6. 计费执行（skipBilling 时整块跳过） ────────────────────

		if !skipBilling {
			// Key 状态检查
			switch apiKey.Status {
			case service.StatusAPIKeyQuotaExhausted:
				if !apiKey.QuotaDisabled {
					AbortWithError(c, 429, "API_KEY_QUOTA_EXHAUSTED", "API key 额度已用完")
					return
				}
			case service.StatusAPIKeyExpired:
				AbortWithError(c, 403, "API_KEY_EXPIRED", "API key 已过期")
				return
			}

			// 运行时过期/配额检查（即使状态是 active，也要检查时间和用量）
			if apiKey.IsExpired() {
				AbortWithError(c, 403, "API_KEY_EXPIRED", "API key 已过期")
				return
			}
			if apiKey.IsQuotaExhausted() {
				AbortWithError(c, 429, "API_KEY_QUOTA_EXHAUSTED", "API key 额度已用完")
				return
			}

			// 订阅模式：验证订阅限额
			if subscription != nil {
				needsMaintenance, validateErr := subscriptionService.ValidateAndCheckLimits(subscription, apiKey.Group)
				if needsMaintenance {
					refreshed, maintenanceErr := subscriptionService.EnsureWindowMaintenance(c.Request.Context(), subscription)
					if maintenanceErr != nil {
						AbortWithError(c, 500, "SUBSCRIPTION_MAINTENANCE_FAILED", "Failed to maintain subscription usage windows")
						return
					}
					subscription = refreshed
					_, validateErr = subscriptionService.ValidateAndCheckLimits(subscription, apiKey.Group)
				}
				if validateErr != nil {
					if isSubscriptionLimitError(validateErr) {
						if automaticSubscriptionResolved {
							alternative, altErr := resolveAlternativeAutomaticSubscriptionEntitlement(c.Request.Context(), subscriptionService, apiKey, entitlementReq, subscription.ID)
							if altErr != nil {
								abortEntitlementResolutionError(c, apiKey, entitlementReq, altErr, "Failed to resolve alternative subscription entitlement")
								return
							}
							if alternative != nil {
								subscription = alternative
								needsMaintenance = false
								validateErr = nil
							}
						}
						if validateErr != nil {
							if ok, fallbackErr := resolveExplicitBalanceFallbackEntitlement(c.Request.Context(), apiKeyService, apiKey, entitlementReq); fallbackErr != nil {
								abortEntitlementResolutionError(c, apiKey, entitlementReq, fallbackErr, "Failed to resolve balance fallback entitlement")
								return
							} else if ok {
								subscription = nil
								if apiKey.User.Balance <= 0 {
									AbortWithError(c, 403, "INSUFFICIENT_BALANCE", "Insufficient account balance")
									return
								}
							} else {
								AbortWithError(c, 429, "USAGE_LIMIT_EXCEEDED", validateErr.Error())
								return
							}
						}
					} else {
						AbortWithError(c, 403, "SUBSCRIPTION_INVALID", validateErr.Error())
						return
					}
				}
				// 窗口维护异步化（不阻塞请求）
				if subscription != nil && needsMaintenance {
					maintenanceCopy := *subscription
					subscriptionService.DoWindowMaintenance(&maintenanceCopy)
				}
			} else {
				// 非订阅模式 或 订阅模式但 subscriptionService 未注入：回退到余额检查
				if apiKeyBalanceBelowAuthThreshold(apiKey.User.Balance, cfg) {
					AbortWithError(c, 403, "INSUFFICIENT_BALANCE", "Insufficient account balance")
					return
				}
			}
		}

		// ── 7. 设置上下文 → Next ─────────────────────────────────────

		setAutomaticSubscriptionResolvedContext(c, automaticSubscriptionResolved)
		if subscription != nil {
			c.Set(string(ContextKeySubscription), subscription)
		}
		c.Set(string(ContextKeyAPIKey), apiKey)
		c.Set(string(ContextKeyUser), AuthSubject{
			UserID:      apiKey.User.ID,
			Concurrency: apiKey.User.Concurrency,
		})
		c.Set(string(ContextKeyUserRole), apiKey.User.Role)
		setGroupContext(c, apiKey.Group)
		_ = apiKeyService.TouchLastUsed(c.Request.Context(), apiKey.ID)

		c.Next()
	}
}

// GetAPIKeyFromContext 从上下文中获取API key
func GetAPIKeyFromContext(c *gin.Context) (*service.APIKey, bool) {
	value, exists := c.Get(string(ContextKeyAPIKey))
	if !exists {
		return nil, false
	}
	apiKey, ok := value.(*service.APIKey)
	return apiKey, ok
}

// SetOpsFallbackAPIKey 记录已加载的 API Key，供 Ops 错误日志在鉴权早退时回退使用。
// 与 ContextKeyAPIKey 区分：写入它不代表请求已通过鉴权，因此不影响 handler、
// 审计日志等对“已鉴权”的判断。
func SetOpsFallbackAPIKey(c *gin.Context, apiKey *service.APIKey) {
	if c == nil || apiKey == nil {
		return
	}
	c.Set(string(ContextKeyOpsFallbackAPIKey), apiKey)
}

// GetOpsFallbackAPIKey 读取 Ops 错误日志专用的回退 API Key。
func GetOpsFallbackAPIKey(c *gin.Context) (*service.APIKey, bool) {
	value, exists := c.Get(string(ContextKeyOpsFallbackAPIKey))
	if !exists {
		return nil, false
	}
	apiKey, ok := value.(*service.APIKey)
	return apiKey, ok
}

// GetSubscriptionFromContext 从上下文中获取订阅信息
func GetSubscriptionFromContext(c *gin.Context) (*service.UserSubscription, bool) {
	value, exists := c.Get(string(ContextKeySubscription))
	if !exists {
		return nil, false
	}
	subscription, ok := value.(*service.UserSubscription)
	return subscription, ok
}

func resolveUnboundAPIKeyEntitlement(ctx context.Context, apiKeyService *service.APIKeyService, subscriptionService *service.SubscriptionService, apiKey *service.APIKey, req service.EntitlementRequest) (*service.UserSubscription, error) {
	if apiKey == nil || apiKey.User == nil {
		return nil, nil
	}

	if apiKey.User.CustomerType != service.CustomerTypeManaged && subscriptionService != nil {
		sub, group, err := subscriptionService.ResolveAutomaticSubscriptionForRequest(ctx, apiKey.User.ID, req)
		if err != nil {
			return nil, err
		}
		if sub != nil && group != nil {
			groupID := group.ID
			apiKey.GroupID = &groupID
			apiKey.Group = group
			sub.Group = group
			return sub, nil
		}
	}

	if _, err := resolveBalanceFallbackEntitlement(ctx, apiKeyService, apiKey, req); err != nil {
		return nil, err
	}

	return nil, nil
}

func resolveAlternativeAutomaticSubscriptionEntitlement(ctx context.Context, subscriptionService *service.SubscriptionService, apiKey *service.APIKey, req service.EntitlementRequest, currentSubscriptionID int64) (*service.UserSubscription, error) {
	if subscriptionService == nil || apiKey == nil || apiKey.User == nil {
		return nil, nil
	}
	candidates, err := subscriptionService.ResolveAutomaticSubscriptionCandidatesForRequest(ctx, apiKey.User.ID, req)
	if err != nil {
		return nil, err
	}
	for i := range candidates {
		candidate := candidates[i]
		if candidate.Subscription.ID == currentSubscriptionID || candidate.Group == nil {
			continue
		}
		groupID := candidate.Group.ID
		apiKey.GroupID = &groupID
		apiKey.Group = candidate.Group
		candidate.Subscription.Group = candidate.Group
		return &candidate.Subscription, nil
	}
	return nil, nil
}

func resolveBalanceFallbackEntitlement(ctx context.Context, apiKeyService *service.APIKeyService, apiKey *service.APIKey, req service.EntitlementRequest) (bool, error) {
	if apiKey == nil || apiKey.User == nil {
		return false, nil
	}
	apiKey.GroupID = nil
	apiKey.Group = nil
	if apiKeyService != nil {
		group, err := apiKeyService.ResolveDefaultStandardGroupForRequest(ctx, apiKey.User.ID, req)
		if err != nil {
			return false, err
		}
		if group != nil {
			groupID := group.ID
			apiKey.GroupID = &groupID
			apiKey.Group = group
		}
	}
	return true, nil
}

func resolveExplicitBalanceFallbackEntitlement(ctx context.Context, apiKeyService *service.APIKeyService, apiKey *service.APIKey, req service.EntitlementRequest) (bool, error) {
	if apiKey == nil || !apiKey.QuotaDisabled {
		return false, nil
	}
	return resolveBalanceFallbackEntitlement(ctx, apiKeyService, apiKey, req)
}

func isSubscriptionLimitError(err error) bool {
	return errors.Is(err, service.ErrDailyLimitExceeded) ||
		errors.Is(err, service.ErrWeeklyLimitExceeded) ||
		errors.Is(err, service.ErrMonthlyLimitExceeded)
}

const entitlementRequestBodyPeekLimit = 128 * 1024

func requestEntitlementRequest(c *gin.Context) service.EntitlementRequest {
	if c == nil || c.Request == nil {
		return service.EntitlementRequest{}
	}
	path := c.Request.URL.Path
	body := peekRequestBody(c)
	model := requestModelFromPathOrBody(path, body)
	requiresImage := service.IsImageGenerationIntent(path, model, body)
	if !requiresImage && bytes.Contains(body, []byte("image_generation")) {
		requiresImage = true
	}
	return service.EntitlementRequest{
		Platform:       inferRequestPlatform(path, model),
		RequestedModel: model,
		RequiresImage:  requiresImage,
	}
}

func peekRequestBody(c *gin.Context) []byte {
	if c == nil || c.Request == nil || c.Request.Body == nil {
		return nil
	}
	switch c.Request.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
	default:
		return nil
	}
	original := c.Request.Body
	buf := make([]byte, entitlementRequestBodyPeekLimit)
	n, err := original.Read(buf)
	if n <= 0 {
		c.Request.Body = original
		return nil
	}
	body := append([]byte(nil), buf[:n]...)
	c.Request.Body = replayReadCloser{
		Reader: io.MultiReader(bytes.NewReader(body), original),
		Closer: original,
	}
	if err != nil && !errors.Is(err, io.EOF) {
		slog.Warn("failed to peek request body for entitlement selection", "path", c.Request.URL.Path, "error", err)
	}
	return body
}

func requestModelFromPathOrBody(path string, body []byte) string {
	if model := geminiModelFromPath(path); model != "" {
		return model
	}
	if len(body) == 0 {
		return ""
	}
	if model := strings.TrimSpace(gjson.GetBytes(body, "model").String()); model != "" {
		return model
	}
	return ""
}

func geminiModelFromPath(path string) string {
	path = strings.TrimSpace(path)
	marker := "/models/"
	idx := strings.Index(path, marker)
	if idx < 0 {
		return ""
	}
	model := path[idx+len(marker):]
	if cut := strings.IndexAny(model, ":/?"); cut >= 0 {
		model = model[:cut]
	}
	return strings.TrimSpace(model)
}

func inferRequestPlatform(path, model string) string {
	lowerPath := strings.ToLower(strings.TrimSpace(path))
	lowerModel := strings.ToLower(strings.TrimSpace(model))
	switch {
	case strings.HasPrefix(lowerPath, "/antigravity/"):
		return service.PlatformAntigravity
	case strings.HasPrefix(lowerPath, "/v1beta/"):
		return service.PlatformGemini
	case strings.Contains(lowerPath, "/images/"),
		strings.Contains(lowerPath, "/responses"),
		strings.Contains(lowerPath, "/chat/completions"):
		return service.PlatformOpenAI
	case strings.HasPrefix(lowerModel, "claude-"):
		return service.PlatformAnthropic
	case strings.HasPrefix(lowerModel, "gemini-"):
		return service.PlatformGemini
	case strings.Contains(lowerModel, "codex"),
		strings.HasPrefix(lowerModel, "gpt-"),
		strings.HasPrefix(lowerModel, "o1"),
		strings.HasPrefix(lowerModel, "o3"),
		strings.HasPrefix(lowerModel, "o4"):
		return service.PlatformOpenAI
	default:
		return ""
	}
}

type replayReadCloser struct {
	io.Reader
	io.Closer
}

func setEntitlementRequestContext(c *gin.Context, req service.EntitlementRequest) {
	if c == nil || c.Request == nil {
		return
	}
	ctx := context.WithValue(c.Request.Context(), ctxkey.SubscriptionEntitlementRequest, req)
	if strings.TrimSpace(req.Platform) != "" {
		ctx = context.WithValue(ctx, ctxkey.Platform, req.Platform)
	}
	if strings.TrimSpace(req.RequestedModel) != "" {
		ctx = context.WithValue(ctx, ctxkey.Model, req.RequestedModel)
	}
	c.Request = c.Request.WithContext(ctx)
}

func setAutomaticSubscriptionResolvedContext(c *gin.Context, resolved bool) {
	if c == nil || c.Request == nil {
		return
	}
	ctx := context.WithValue(c.Request.Context(), ctxkey.AutomaticSubscriptionResolved, resolved)
	c.Request = c.Request.WithContext(ctx)
}

func abortEntitlementResolutionError(c *gin.Context, apiKey *service.APIKey, req service.EntitlementRequest, err error, message string) {
	status, code, responseMessage := classifyEntitlementResolutionError(c, err, message)
	logEntitlementResolutionError(c, apiKey, req, err, status, message)
	AbortWithError(c, status, code, responseMessage)
}

func classifyEntitlementResolutionError(c *gin.Context, err error, message string) (int, string, string) {
	if requestContextDone(c, err) {
		return 499, "CLIENT_CLOSED_REQUEST", "Client closed request"
	}
	return http.StatusInternalServerError, "INTERNAL_ERROR", message
}

func requestContextDone(c *gin.Context, err error) bool {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	if c == nil || c.Request == nil {
		return false
	}
	ctxErr := c.Request.Context().Err()
	return errors.Is(ctxErr, context.Canceled) || errors.Is(ctxErr, context.DeadlineExceeded)
}

func logEntitlementResolutionError(c *gin.Context, apiKey *service.APIKey, req service.EntitlementRequest, err error, status int, message string) {
	attrs := []any{
		"status", status,
		"message", message,
		"error", err,
		"platform", req.Platform,
		"requested_model", req.RequestedModel,
		"requires_image", req.RequiresImage,
	}
	if c != nil && c.Request != nil {
		attrs = append(attrs, "path", c.Request.URL.Path, "request_context_error", c.Request.Context().Err())
	}
	if apiKey != nil {
		attrs = append(attrs, "api_key_id", apiKey.ID)
		if apiKey.GroupID != nil {
			attrs = append(attrs, "group_id", *apiKey.GroupID)
		}
		if apiKey.User != nil {
			attrs = append(attrs, "user_id", apiKey.User.ID)
		}
	}
	slog.Warn("api_key_entitlement_resolution_failed", attrs...)
}

func setGroupContext(c *gin.Context, group *service.Group) {
	if !service.IsGroupContextValid(group) {
		return
	}
	if existing, ok := c.Request.Context().Value(ctxkey.Group).(*service.Group); ok && existing != nil && existing.ID == group.ID && service.IsGroupContextValid(existing) {
		return
	}
	ctx := context.WithValue(c.Request.Context(), ctxkey.Group, group)
	c.Request = c.Request.WithContext(ctx)
}

// apiKeyBalanceBelowAuthThreshold 保持鉴权层的历史语义：仅在余额耗尽（<=0）时拒绝。
// MinimumBalanceReserve 只作为 billing-cache 预检的保守下限，不得复用为鉴权硬门槛，
// 否则已配置该值的存量部署升级后，0 < balance < reserve 的用户会在所有端点被静默 403。
func apiKeyBalanceBelowAuthThreshold(balance float64, _ *config.Config) bool {
	return balance <= 0
}

func abortIfAPIKeyGroupUnavailable(c *gin.Context, apiKey *service.APIKey) bool {
	code, message, ok := validateAPIKeyGroupAvailable(apiKey)
	if ok {
		return false
	}
	service.MarkOpsClientBusinessLimited(c, service.OpsClientBusinessLimitedReasonAPIKeyGroupUnavailable)
	AbortWithError(c, 403, code, message)
	return true
}

func abortIfAPIKeyGroupNotAllowed(c *gin.Context, apiKey *service.APIKey) bool {
	if validateAPIKeyGroupAllowed(apiKey) {
		return false
	}
	service.MarkOpsClientBusinessLimited(c, service.OpsClientBusinessLimitedReasonAPIKeyGroupUnavailable)
	AbortWithError(c, 403, "GROUP_NOT_ALLOWED", "API Key 所属专属分组不再允许当前用户使用")
	return true
}

func validateAPIKeyGroupAllowed(apiKey *service.APIKey) bool {
	if apiKey == nil || apiKey.GroupID == nil || apiKey.User == nil || apiKey.Group == nil {
		return true
	}
	group := apiKey.Group
	if group.IsSubscriptionType() {
		return true
	}
	return apiKey.User.CanBindGroup(group.ID, group.IsExclusive)
}

func validateAPIKeyGroupAvailable(apiKey *service.APIKey) (string, string, bool) {
	if apiKey == nil || apiKey.GroupID == nil {
		return "", "", true
	}
	group := apiKey.Group
	if group == nil || strings.EqualFold(group.Status, "deleted") {
		return "GROUP_DELETED", "API Key 所属分组已删除", false
	}
	if !group.IsActive() {
		return "GROUP_DISABLED", "API Key 所属分组已停用", false
	}
	return "", "", true
}
