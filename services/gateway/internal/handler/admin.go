package handler

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/KeKe-Li/ai-token/services/gateway/internal/model"
	"github.com/KeKe-Li/ai-token/services/gateway/internal/service"
)

type AdminHandler struct {
	channelStore  *model.ChannelStore
	modelStore    *model.ModelStore
	userStore     *model.UserStore
	logStore      *model.LogStore
	encryptionKey string
}

func NewAdminHandler(channelStore *model.ChannelStore, modelStore *model.ModelStore, userStore *model.UserStore, logStore *model.LogStore, encryptionKey string) *AdminHandler {
	return &AdminHandler{
		channelStore:  channelStore,
		modelStore:    modelStore,
		userStore:     userStore,
		logStore:      logStore,
		encryptionKey: encryptionKey,
	}
}

// ===== 渠道管理 =====

func (h *AdminHandler) ListChannels(c *gin.Context) {
	channels, err := h.channelStore.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取渠道列表失败"})
		return
	}
	if channels == nil {
		channels = []model.Channel{}
	}
	c.JSON(http.StatusOK, gin.H{"data": channels})
}

type createChannelRequest struct {
	Name      string   `json:"name" binding:"required"`
	Provider  string   `json:"provider" binding:"required"`
	BaseURL   string   `json:"base_url" binding:"required"`
	APIKey    string   `json:"api_key" binding:"required"`
	Models    []string `json:"models" binding:"required"`
	Priority  int      `json:"priority"`
	Weight    int      `json:"weight"`
	RateLimit int      `json:"rate_limit"`
}

func (h *AdminHandler) CreateChannel(c *gin.Context) {
	var req createChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	if req.Weight <= 0 {
		req.Weight = 1
	}

	ch := &model.Channel{
		Name:      req.Name,
		Provider:  req.Provider,
		BaseURL:   req.BaseURL,
		APIKeyEnc: h.encryptAPIKey(req.APIKey),
		Models:    req.Models,
		Status:    1,
		Priority:  req.Priority,
		Weight:    req.Weight,
		RateLimit: req.RateLimit,
	}

	if err := h.channelStore.Create(c.Request.Context(), ch); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建渠道失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": ch})
}

func (h *AdminHandler) UpdateChannel(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 ID"})
		return
	}

	existing, err := h.channelStore.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "渠道不存在"})
		return
	}

	var req createChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	existing.Name = req.Name
	existing.Provider = req.Provider
	existing.BaseURL = req.BaseURL
	existing.Models = req.Models
	existing.Priority = req.Priority
	existing.Weight = req.Weight
	existing.RateLimit = req.RateLimit
	if req.APIKey != "" {
		existing.APIKeyEnc = h.encryptAPIKey(req.APIKey)
	}

	if err := h.channelStore.Update(c.Request.Context(), existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新渠道失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": existing})
}

func (h *AdminHandler) DeleteChannel(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 ID"})
		return
	}

	if err := h.channelStore.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "渠道不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "渠道已删除"})
}

func (h *AdminHandler) TestChannel(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 ID"})
		return
	}

	var req struct {
		Model string `json:"model"`
	}
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	ch, err := h.channelStore.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "渠道不存在"})
		return
	}

	apiKey, err := service.Decrypt(ch.APIKeyEnc, h.encryptionKey)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"ok":      false,
			"status":  0,
			"message": "渠道密钥解密失败，请重新保存渠道密钥",
		})
		return
	}

	ok, statusCode, latency, message := service.ProbeChannelDetailed(ch.Provider, ch.BaseURL, apiKey)
	c.JSON(http.StatusOK, gin.H{
		"ok":         ok,
		"status":     statusCode,
		"latency_ms": latency,
		"message":    message,
		"model":      req.Model,
	})
}

// ===== 模型管理 =====

func (h *AdminHandler) ListModels(c *gin.Context) {
	models, err := h.modelStore.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取模型列表失败"})
		return
	}
	if models == nil {
		models = []model.AIModel{}
	}
	c.JSON(http.StatusOK, gin.H{"data": models})
}

type createModelRequest struct {
	ModelID       string   `json:"model_id" binding:"required"`
	DisplayName   string   `json:"display_name" binding:"required"`
	Provider      string   `json:"provider" binding:"required"`
	Category      string   `json:"category" binding:"required"`
	ContextLength int      `json:"context_length"`
	InputPrice    int64    `json:"input_price"`
	OutputPrice   int64    `json:"output_price"`
	PriceUnit     string   `json:"price_unit"`
	Capabilities  []string `json:"capabilities"`
	Description   *string  `json:"description"`
}

func (h *AdminHandler) CreateModel(c *gin.Context) {
	var req createModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	if req.PriceUnit == "" {
		req.PriceUnit = "1M"
	}

	m := &model.AIModel{
		ModelID:       req.ModelID,
		DisplayName:   req.DisplayName,
		Provider:      req.Provider,
		Category:      req.Category,
		ContextLength: req.ContextLength,
		InputPrice:    req.InputPrice,
		OutputPrice:   req.OutputPrice,
		PriceUnit:     req.PriceUnit,
		Capabilities:  req.Capabilities,
		Status:        1,
		Description:   req.Description,
	}

	if err := h.modelStore.Create(c.Request.Context(), m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建模型失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": m})
}

func (h *AdminHandler) UpdateModel(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 ID"})
		return
	}

	var req createModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	if req.PriceUnit == "" {
		req.PriceUnit = "1M"
	}

	m := &model.AIModel{
		ID:            id,
		DisplayName:   req.DisplayName,
		Provider:      req.Provider,
		Category:      req.Category,
		ContextLength: req.ContextLength,
		InputPrice:    req.InputPrice,
		OutputPrice:   req.OutputPrice,
		PriceUnit:     req.PriceUnit,
		Capabilities:  req.Capabilities,
		Status:        1,
		Description:   req.Description,
	}

	if err := h.modelStore.Update(c.Request.Context(), m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新模型失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": m})
}

func (h *AdminHandler) DeleteModel(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 ID"})
		return
	}

	if err := h.modelStore.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "模型不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "模型已删除"})
}

// ===== 用户管理 =====

func (h *AdminHandler) ListUsers(c *gin.Context) {
	// 简化实现：直接查询
	rows, err := h.userStore.ListAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户列表失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rows})
}

func (h *AdminHandler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 ID"})
		return
	}

	var req struct {
		Status  *int   `json:"status"`
		Balance *int64 `json:"balance"`
		Role    *int   `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if err := h.userStore.AdminUpdate(c.Request.Context(), id, req.Status, req.Balance, req.Role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新用户失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "用户已更新"})
}

// ===== 全局日志和统计 =====

func (h *AdminHandler) GlobalLogs(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if limit > 100 {
		limit = 100
	}

	logs, err := h.logStore.ListAll(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取日志失败"})
		return
	}
	if logs == nil {
		logs = []model.RequestLog{}
	}
	c.JSON(http.StatusOK, gin.H{"data": logs})
}

func (h *AdminHandler) Stats(c *gin.Context) {
	since := time.Now().AddDate(0, 0, -1)

	stats, err := h.logStore.GetGlobalStats(c.Request.Context(), since)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取统计失败"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *AdminHandler) encryptAPIKey(raw string) string {
	encrypted, err := service.Encrypt(raw, h.encryptionKey)
	if err != nil {
		return raw
	}
	return encrypted
}
