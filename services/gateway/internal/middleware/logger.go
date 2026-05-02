package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)

		status := c.Writer.Status()
		if status >= 400 {
			log.Printf("[%d] %s %s %v | %s | %s",
				status, c.Request.Method, c.Request.URL.Path,
				latency, c.ClientIP(), c.GetString("request_id"))
		}
	}
}
