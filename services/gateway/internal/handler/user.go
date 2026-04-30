package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/KeKe-Li/ai-token/services/gateway/internal/model"
)

type UserHandler struct {
	userStore   *model.UserStore
	apiKeyStore *model.APIKeyStore
	logStore    *model.LogStore
}

func NewUserHandler(userStore *model.UserStore, apiKeyStore *model.APIKeyStore, logStore *model.LogStore) *UserHandler {
	return &UserHandler{
		userStore:   userStore,
		apiKeyStore: apiKeyStore,
		logStore:    logStore,
	}
}

func (h *UserHandler) Profile(c *gin.Context) {
	userID := c.GetInt64("user_id")
	user, err := h.userStore.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":            user.ID,
		"username":      user.Username,
		"email":         user.Email,
		"role":          user.Role,
		"balance":       user.Balance,
		"used_amount":   user.UsedAmount,
		"request_count": user.RequestCount,
		"group_name":    user.GroupName,
		"created_at":    user.CreatedAt,
	})
}

func (h *UserHandler) Dashboard(c *gin.Context) {
	userID := c.GetInt64("user_id")
	user, err := h.userStore.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	since := time.Now().AddDate(0, -1, 0)
	usage, _ := h.logStore.GetUsageSummary(c.Request.Context(), userID, since)

	c.JSON(http.StatusOK, gin.H{
		"balance":       user.Balance,
		"used_amount":   user.UsedAmount,
		"request_count": user.RequestCount,
		"usage":         usage,
	})
}

func (h *UserHandler) ListKeys(c *gin.Context) {
	userID := c.GetInt64("user_id")
	keys, err := h.apiKeyStore.ListByUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取密钥列表失败"})
		return
	}
	if keys == nil {
		keys = []model.APIKey{}
	}
	c.JSON(http.StatusOK, gin.H{"data": keys})
}

type createKeyRequest struct {
	Name   string   `json:"name" binding:"required"`
	Models []string `json:"models"`
}

func (h *UserHandler) CreateKey(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req createKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	apiKey, rawKey, err := h.apiKeyStore.Create(c.Request.Context(), userID, req.Name, req.Models)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建密钥失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"key":  rawKey,
		"data": apiKey,
		"message": "密钥已创建，请立即保存，此后无法再次查看完整密钥",
	})
}

func (h *UserHandler) DeleteKey(c *gin.Context) {
	userID := c.GetInt64("user_id")
	keyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 ID"})
		return
	}

	if err := h.apiKeyStore.Delete(c.Request.Context(), keyID, userID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "密钥不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "密钥已删除"})
}

func (h *UserHandler) Logs(c *gin.Context) {
	userID := c.GetInt64("user_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	logs, err := h.logStore.ListByUser(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取日志失败"})
		return
	}
	if logs == nil {
		logs = []model.RequestLog{}
	}

	c.JSON(http.StatusOK, gin.H{"data": logs})
}

func (h *UserHandler) Usage(c *gin.Context) {
	userID := c.GetInt64("user_id")
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	since := time.Now().AddDate(0, 0, -days)

	usage, err := h.logStore.GetUsageSummary(c.Request.Context(), userID, since)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用量统计失败"})
		return
	}
	if usage == nil {
		usage = []model.UsageSummary{}
	}

	c.JSON(http.StatusOK, gin.H{"data": usage})
}
