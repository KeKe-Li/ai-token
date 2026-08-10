package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/KeKe-Li/ai-token/services/gateway/internal/model"
)

// publicCacheControl 是公开模型接口的缓存策略：模型数据几乎静态，
// 允许客户端与中间层缓存 5 分钟，降低无鉴权接口对数据库的回源压力。
const publicCacheControl = "public, max-age=300"

type PublicHandler struct {
	modelStore *model.ModelStore
}

func NewPublicHandler(modelStore *model.ModelStore) *PublicHandler {
	return &PublicHandler{modelStore: modelStore}
}

func (h *PublicHandler) ListModels(c *gin.Context) {
	models, err := h.modelStore.ListActive(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取模型列表失败"})
		return
	}
	if models == nil {
		models = []model.AIModel{}
	}
	c.Header("Cache-Control", publicCacheControl)
	c.JSON(http.StatusOK, gin.H{"data": models})
}

func (h *PublicHandler) GetModel(c *gin.Context) {
	// 路由为 catch-all (*id)，含斜杠的 model_id（如 deepseek/deepseek-v4-pro）
	// 会带上前导斜杠，需去除后还原完整 id。
	modelID := strings.TrimPrefix(c.Param("id"), "/")
	if modelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少模型 ID"})
		return
	}

	m, err := h.modelStore.GetByModelID(c.Request.Context(), modelID)
	if err != nil {
		if errors.Is(err, model.ErrModelNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "模型不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取模型失败"})
		return
	}
	c.Header("Cache-Control", publicCacheControl)
	c.JSON(http.StatusOK, gin.H{"data": m})
}
