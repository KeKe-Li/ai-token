package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/KeKe-Li/ai-token/services/gateway/internal/model"
	"github.com/KeKe-Li/ai-token/services/gateway/internal/relay"
	"github.com/KeKe-Li/ai-token/services/gateway/internal/relay/adaptor"
)

type RelayHandler struct {
	engine     *relay.RelayEngine
	channels   []relay.Channel
	modelStore *model.ModelStore
}

func NewRelayHandler(engine *relay.RelayEngine, modelStore *model.ModelStore) *RelayHandler {
	return &RelayHandler{
		engine:     engine,
		modelStore: modelStore,
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
	for _, ch := range h.channels {
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

	resp, channel, err := h.engine.ChatCompletion(c.Request.Context(), h.channels, req, 2)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}

	latency := int(time.Since(start).Milliseconds())

	h.engine.RecordUsage(relay.UsageRecord{
		UserID:       c.GetInt64("user_id"),
		APIKeyID:     c.GetInt64("api_key_id"),
		ChannelID:    channel.ID,
		Model:        req.Model,
		Method:       "POST",
		Path:         "/v1/completions",
		StatusCode:   200,
		InputTokens:  resp.Usage.PromptTokens,
		OutputTokens: resp.Usage.CompletionTokens,
		LatencyMs:    latency,
		IP:           c.ClientIP(),
	})

	c.JSON(http.StatusOK, resp)
}
