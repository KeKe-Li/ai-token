package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/KeKe-Li/ai-token/services/gateway/internal/model"
)

func APIKeyAuth(keyStore *model.APIKeyStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"message": "缺少 Authorization 头",
					"type":    "authentication_error",
				},
			})
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader || !strings.HasPrefix(token, "sk-") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"message": "无效的 API Key 格式",
					"type":    "authentication_error",
				},
			})
			c.Abort()
			return
		}

		apiKey, err := keyStore.ValidateKey(c.Request.Context(), token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"message": "无效的 API Key",
					"type":    "authentication_error",
				},
			})
			c.Abort()
			return
		}

		c.Set("api_key_id", apiKey.ID)
		c.Set("user_id", apiKey.UserID)
		c.Set("allowed_models", apiKey.Models)
		c.Next()
	}
}
