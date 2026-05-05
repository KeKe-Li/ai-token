package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/KeKe-Li/ai-token/services/gateway/internal/model"
	"github.com/KeKe-Li/ai-token/services/gateway/internal/relay"
	"github.com/KeKe-Li/ai-token/services/gateway/internal/relay/adaptor"
	"github.com/KeKe-Li/ai-token/services/gateway/internal/service"
)

type RelayHandler struct {
	engine          *relay.RelayEngine
	envChannels     []relay.Channel
	channelStore    *model.ChannelStore
	modelStore      *model.ModelStore
	userStore       *model.UserStore
	pricingProvider relay.PricingProvider
	encryptionKey   string
}

type preAuthorizationResult struct {
	ReservedAmount int64
	WalletHoldID   int64
}

func NewRelayHandler(engine *relay.RelayEngine, modelStore *model.ModelStore, channelStore *model.ChannelStore, userStore *model.UserStore, pricingProvider relay.PricingProvider, encryptionKey string) *RelayHandler {
	return &RelayHandler{
		engine:          engine,
		modelStore:      modelStore,
		channelStore:    channelStore,
		userStore:       userStore,
		pricingProvider: pricingProvider,
		encryptionKey:   encryptionKey,
	}
}

func (h *RelayHandler) SetChannels(channels []relay.Channel) {
	h.envChannels = channels
}

func (h *RelayHandler) loadChannels(ctx context.Context) []relay.Channel {
	if h.channelStore != nil {
		dbChannels, err := h.channelStore.List(ctx)
		if err == nil && len(dbChannels) > 0 {
			channels := make([]relay.Channel, 0, len(dbChannels))
			for _, ch := range dbChannels {
				relayChannel, err := service.BuildRelayChannel(ch, h.encryptionKey)
				if err != nil {
					log.Printf("跳过无效数据库渠道 #%d (%s): %v", ch.ID, ch.Name, err)
					continue
				}
				channels = append(channels, relayChannel)
			}
			return channels
		}
	}
	return h.envChannels
}

func (h *RelayHandler) ChatCompletions(c *gin.Context) {
	var req adaptor.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "请求格式错误: " + err.Error(),
				"type":    "invalid_request_error",
			},
		})
		return
	}

	if req.Model == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "model 字段为必填",
				"type":    "invalid_request_error",
			},
		})
		return
	}

	if len(req.Messages) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "messages 不能为空",
				"type":    "invalid_request_error",
			},
		})
		return
	}

	if !h.isRequestModelAllowed(c, req.Model) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": gin.H{
				"message": "API Key 不允许访问模型 " + req.Model,
				"type":    "permission_error",
				"code":    "model_not_allowed",
			},
		})
		return
	}

	preAuth, ok := h.ensurePreAuthorization(c, &req)
	if !ok {
		return
	}

	start := time.Now()

	if req.Stream {
		h.handleStream(c, &req, start, preAuth)
		return
	}

	h.handleNonStream(c, &req, start, preAuth)
}

func (h *RelayHandler) handleNonStream(c *gin.Context, req *adaptor.ChatRequest, start time.Time, preAuth preAuthorizationResult) {
	resp, channel, err := h.engine.ChatCompletion(c.Request.Context(), h.loadChannels(c.Request.Context()), req, 2)
	if err != nil {
		h.releasePreAuthorization(c, preAuth, "上游非流式调用失败，释放预授权 hold")
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{
				"message": "上游服务调用失败: " + err.Error(),
				"type":    "upstream_error",
			},
		})
		return
	}

	latency := int(time.Since(start).Milliseconds())

	h.engine.RecordUsage(relay.UsageRecord{
		UserID:         c.GetInt64("user_id"),
		APIKeyID:       c.GetInt64("api_key_id"),
		ChannelID:      channel.ID,
		Model:          req.Model,
		Method:         "POST",
		Path:           "/v1/chat/completions",
		StatusCode:     200,
		InputTokens:    resp.Usage.PromptTokens,
		OutputTokens:   resp.Usage.CompletionTokens,
		ReservedAmount: preAuth.ReservedAmount,
		WalletHoldID:   preAuth.WalletHoldID,
		LatencyMs:      latency,
		IP:             c.ClientIP(),
	})

	c.JSON(http.StatusOK, resp)
}

func (h *RelayHandler) handleStream(c *gin.Context, req *adaptor.ChatRequest, start time.Time, preAuth preAuthorizationResult) {
	stream, channel, err := h.engine.ChatCompletionStream(c.Request.Context(), h.loadChannels(c.Request.Context()), req, 2)
	if err != nil {
		h.releasePreAuthorization(c, preAuth, "上游流式调用建立失败，释放预授权 hold")
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{
				"message": "上游服务调用失败: " + err.Error(),
				"type":    "upstream_error",
			},
		})
		return
	}

	streamResult := relay.WriteStreamResponse(c, stream)
	latency := int(time.Since(start).Milliseconds())

	h.engine.RecordUsage(relay.UsageRecord{
		UserID:          c.GetInt64("user_id"),
		APIKeyID:        c.GetInt64("api_key_id"),
		ChannelID:       channel.ID,
		Model:           req.Model,
		Method:          "POST",
		Path:            "/v1/chat/completions",
		StatusCode:      200,
		InputTokens:     streamResult.InputTokens,
		OutputTokens:    streamResult.OutputTokens,
		ReservedAmount:  preAuth.ReservedAmount,
		WalletHoldID:    preAuth.WalletHoldID,
		LatencyMs:       latency,
		IP:              c.ClientIP(),
		EstimatedTokens: streamResult.Estimated,
	})
}

func (h *RelayHandler) ListModels(c *gin.Context) {
	type modelObject struct {
		ID      string `json:"id"`
		Object  string `json:"object"`
		Created int64  `json:"created"`
		OwnedBy string `json:"owned_by"`
	}

	// 优先从数据库读取
	if h.modelStore != nil {
		dbModels, err := h.modelStore.List(c.Request.Context())
		if err == nil && len(dbModels) > 0 {
			result := make([]modelObject, len(dbModels))
			for i, m := range dbModels {
				result[i] = modelObject{
					ID:      m.ModelID,
					Object:  "model",
					Created: m.CreatedAt.Unix(),
					OwnedBy: m.Provider,
				}
			}
			c.JSON(http.StatusOK, gin.H{"object": "list", "data": result})
			return
		}
	}

	// 降级：从渠道配置中聚合所有模型
	seen := make(map[string]bool)
	var result []modelObject
	for _, ch := range h.loadChannels(c.Request.Context()) {
		for _, m := range ch.Models {
			if !seen[m] {
				seen[m] = true
				result = append(result, modelObject{
					ID:      m,
					Object:  "model",
					Created: 1700000000,
					OwnedBy: ch.Provider,
				})
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"object": "list", "data": result})
}

func (h *RelayHandler) Completions(c *gin.Context) {
	start := time.Now()

	var body map[string]any
	if err := json.NewDecoder(c.Request.Body).Decode(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid request"}})
		return
	}

	prompt, _ := body["prompt"].(string)
	model, _ := body["model"].(string)

	req := &adaptor.ChatRequest{
		Model:    model,
		Messages: []adaptor.Message{{Role: "user", Content: prompt}},
	}
	if rawMaxTokens, ok := body["max_tokens"].(float64); ok {
		maxTokens := int(rawMaxTokens)
		req.MaxTokens = &maxTokens
	}

	if !h.isRequestModelAllowed(c, req.Model) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": gin.H{
				"message": "API Key 不允许访问模型 " + req.Model,
				"type":    "permission_error",
				"code":    "model_not_allowed",
			},
		})
		return
	}

	preAuth, ok := h.ensurePreAuthorization(c, req)
	if !ok {
		return
	}

	resp, channel, err := h.engine.ChatCompletion(c.Request.Context(), h.loadChannels(c.Request.Context()), req, 2)
	if err != nil {
		h.releasePreAuthorization(c, preAuth, "上游 completions 调用失败，释放预授权 hold")
		c.JSON(http.StatusBadGateway, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}

	latency := int(time.Since(start).Milliseconds())

	h.engine.RecordUsage(relay.UsageRecord{
		UserID:         c.GetInt64("user_id"),
		APIKeyID:       c.GetInt64("api_key_id"),
		ChannelID:      channel.ID,
		Model:          req.Model,
		Method:         "POST",
		Path:           "/v1/completions",
		StatusCode:     200,
		InputTokens:    resp.Usage.PromptTokens,
		OutputTokens:   resp.Usage.CompletionTokens,
		ReservedAmount: preAuth.ReservedAmount,
		WalletHoldID:   preAuth.WalletHoldID,
		LatencyMs:      latency,
		IP:             c.ClientIP(),
	})

	c.JSON(http.StatusOK, resp)
}

func (h *RelayHandler) isRequestModelAllowed(c *gin.Context, model string) bool {
	raw, exists := c.Get("allowed_models")
	if !exists || raw == nil {
		return true
	}
	allowed, ok := raw.([]string)
	if !ok {
		return true
	}
	return modelAllowed(model, allowed)
}

func modelAllowed(model string, allowed []string) bool {
	if len(allowed) == 0 {
		return true
	}
	for _, item := range allowed {
		pattern := strings.TrimSpace(item)
		if pattern == "" {
			continue
		}
		if pattern == "*" || pattern == model {
			return true
		}
		if strings.HasSuffix(pattern, "*") && strings.HasPrefix(model, strings.TrimSuffix(pattern, "*")) {
			return true
		}
	}
	return false
}

func (h *RelayHandler) ensurePreAuthorization(c *gin.Context, req *adaptor.ChatRequest) (preAuthorizationResult, bool) {
	if h.userStore == nil || h.pricingProvider == nil {
		return preAuthorizationResult{}, true
	}

	pricing, status, note := relay.ResolvePricing(c.Request.Context(), h.pricingProvider, req.Model)
	if status == relay.BillingStatusUnpriced {
		h.recordBillingEvent(c, model.BillingEventModelUnpriced, req.Model, 0, model.WalletHoldStatusFailed, note)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "模型未配置价格，无法发起计费请求: " + req.Model,
				"type":    "billing_error",
				"code":    "model_unpriced",
				"detail":  note,
			},
		})
		return preAuthorizationResult{}, false
	}

	reservedAmount := relay.EstimatePreAuthorizationCost(pricing, req.MaxTokens, 1024)
	if reservedAmount <= 0 {
		return preAuthorizationResult{}, true
	}

	userID := c.GetInt64("user_id")
	holdResult, err := h.userStore.CreateWalletHold(c.Request.Context(), userID, c.GetInt64("api_key_id"), req.Model, reservedAmount)
	if err != nil {
		h.recordBillingEvent(c, model.BillingEventPreAuthError, req.Model, reservedAmount, model.WalletHoldStatusFailed, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": "余额预授权检查失败",
				"type":    "server_error",
			},
		})
		return preAuthorizationResult{}, false
	}

	decision := holdResult.Decision
	if !decision.Allowed {
		c.JSON(http.StatusPaymentRequired, gin.H{
			"error": gin.H{
				"message": "余额不足，低于本次请求预授权金额",
				"type":    "insufficient_balance",
				"code":    relay.BillingStatusPreAuthFailed,
			},
			"required_amount": decision.RequiredAmount,
			"balance":         decision.Balance,
			"reserved":        decision.Reserved,
			"available":       decision.Available,
		})
		return preAuthorizationResult{}, false
	}

	return preAuthorizationResult{ReservedAmount: reservedAmount, WalletHoldID: holdResult.HoldID}, true
}

func (h *RelayHandler) releasePreAuthorization(c *gin.Context, preAuth preAuthorizationResult, note string) {
	if h.userStore == nil || preAuth.WalletHoldID <= 0 {
		return
	}
	if err := h.userStore.ReleaseWalletHold(c.Request.Context(), preAuth.WalletHoldID, note); err != nil {
		log.Printf("释放预授权 hold #%d 失败: %v", preAuth.WalletHoldID, err)
	}
}

func (h *RelayHandler) recordBillingEvent(c *gin.Context, eventType, modelID string, amount int64, status string, note string) {
	if h.userStore == nil {
		return
	}
	if err := h.userStore.CreateBillingEvent(c.Request.Context(), model.BillingEvent{
		UserID:    c.GetInt64("user_id"),
		APIKeyID:  c.GetInt64("api_key_id"),
		EventType: eventType,
		Model:     modelID,
		Amount:    amount,
		Status:    status,
		Note:      note,
	}); err != nil {
		log.Printf("写入账务事件失败: %v", err)
	}
}
