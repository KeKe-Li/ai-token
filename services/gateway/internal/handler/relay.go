package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/KeKe-Li/ai-token/services/gateway/internal/relay"
	"github.com/KeKe-Li/ai-token/services/gateway/internal/relay/adaptor"
)

type RelayHandler struct {
	engine   *relay.RelayEngine
	channels []relay.Channel
}

func NewRelayHandler(engine *relay.RelayEngine) *RelayHandler {
	return &RelayHandler{
		engine: engine,
	}
}

func (h *RelayHandler) SetChannels(channels []relay.Channel) {
	h.channels = channels
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

	start := time.Now()

	if req.Stream {
		h.handleStream(c, &req, start)
		return
	}

	h.handleNonStream(c, &req, start)
}

func (h *RelayHandler) handleNonStream(c *gin.Context, req *adaptor.ChatRequest, start time.Time) {
	resp, channel, err := h.engine.ChatCompletion(c.Request.Context(), h.channels, req, 2)
	if err != nil {
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
		UserID:       c.GetInt64("user_id"),
		APIKeyID:     c.GetInt64("api_key_id"),
		ChannelID:    channel.ID,
		Model:        req.Model,
		Method:       "POST",
		Path:         "/v1/chat/completions",
		StatusCode:   200,
		InputTokens:  resp.Usage.PromptTokens,
		OutputTokens: resp.Usage.CompletionTokens,
		LatencyMs:    latency,
		IP:           c.ClientIP(),
	})

	c.JSON(http.StatusOK, resp)
}

func (h *RelayHandler) handleStream(c *gin.Context, req *adaptor.ChatRequest, start time.Time) {
	stream, channel, err := h.engine.ChatCompletionStream(c.Request.Context(), h.channels, req, 2)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{
				"message": "上游服务调用失败: " + err.Error(),
				"type":    "upstream_error",
			},
		})
		return
	}

	chunkCount, _ := relay.WriteStreamResponse(c, stream)
	latency := int(time.Since(start).Milliseconds())

	h.engine.RecordUsage(relay.UsageRecord{
		UserID:       c.GetInt64("user_id"),
		APIKeyID:     c.GetInt64("api_key_id"),
		ChannelID:    channel.ID,
		Model:        req.Model,
		Method:       "POST",
		Path:         "/v1/chat/completions",
		StatusCode:   200,
		OutputTokens: chunkCount,
		LatencyMs:    latency,
		IP:           c.ClientIP(),
	})
}

func (h *RelayHandler) ListModels(c *gin.Context) {
	type modelObject struct {
		ID      string `json:"id"`
		Object  string `json:"object"`
		Created int64  `json:"created"`
		OwnedBy string `json:"owned_by"`
	}

	models := []modelObject{
		{ID: "gpt-4o", Object: "model", Created: 1700000000, OwnedBy: "openai"},
		{ID: "gpt-4o-mini", Object: "model", Created: 1700000000, OwnedBy: "openai"},
		{ID: "claude-sonnet-4-6", Object: "model", Created: 1700000000, OwnedBy: "anthropic"},
		{ID: "claude-haiku-4-5", Object: "model", Created: 1700000000, OwnedBy: "anthropic"},
		{ID: "gemini-2.5-pro", Object: "model", Created: 1700000000, OwnedBy: "google"},
		{ID: "gemini-2.5-flash", Object: "model", Created: 1700000000, OwnedBy: "google"},
		{ID: "deepseek-chat", Object: "model", Created: 1700000000, OwnedBy: "deepseek"},
		{ID: "deepseek-reasoner", Object: "model", Created: 1700000000, OwnedBy: "deepseek"},
	}

	c.JSON(http.StatusOK, gin.H{
		"object": "list",
		"data":   models,
	})
}

func (h *RelayHandler) Completions(c *gin.Context) {
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

	resp, _, err := h.engine.ChatCompletion(c.Request.Context(), h.channels, req, 2)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, resp)
}
