package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func BalanceCheck(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			c.Next()
			return
		}

		userID := c.GetInt64("user_id")
		if userID == 0 {
			c.Next()
			return
		}

		var balance int64
		err := db.QueryRow(c.Request.Context(), `SELECT balance FROM users WHERE id = $1`, userID).Scan(&balance)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{"message": "余额查询失败", "type": "server_error"},
			})
			c.Abort()
			return
		}

		if balance <= 0 {
			c.JSON(http.StatusPaymentRequired, gin.H{
				"error": gin.H{"message": "余额不足，请充值后重试", "type": "insufficient_balance"},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
