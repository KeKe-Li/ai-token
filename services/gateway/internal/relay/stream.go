package relay

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/KeKe-Li/ai-token/services/gateway/internal/relay/adaptor"
)

func WriteStreamResponse(c *gin.Context, stream <-chan adaptor.StreamChunk) (int, int) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming not supported"})
		return 0, 0
	}

	totalOutput := 0
	chunkCount := 0

	for chunk := range stream {
		if chunk.Error != nil {
			errData := fmt.Sprintf("data: {\"error\": \"%s\"}\n\n", chunk.Error.Error())
			c.Writer.Write([]byte(errData))
			flusher.Flush()
			break
		}

		if chunk.Done {
			c.Writer.Write([]byte("data: [DONE]\n\n"))
			flusher.Flush()
			break
		}

		if len(chunk.Data) > 0 {
			c.Writer.Write(chunk.Data)
			flusher.Flush()
			chunkCount++
			totalOutput += len(chunk.Data)
		}
	}

	return chunkCount, totalOutput
}
