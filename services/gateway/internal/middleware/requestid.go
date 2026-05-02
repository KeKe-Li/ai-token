package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := generateRequestID()
		c.Set("request_id", id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}

func generateRequestID() string {
	ts := time.Now().UnixMilli()
	b := make([]byte, 4)
	rand.Read(b)
	return fmt.Sprintf("req-%x-%s", ts, hex.EncodeToString(b))
}
