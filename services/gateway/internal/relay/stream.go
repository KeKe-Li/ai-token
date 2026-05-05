package relay

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/KeKe-Li/ai-token/services/gateway/internal/relay/adaptor"
)

type StreamResult struct {
	InputTokens  int
	OutputTokens int
	Estimated    bool
}

func WriteStreamResponse(c *gin.Context, stream <-chan adaptor.StreamChunk) StreamResult {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming not supported"})
		return StreamResult{}
	}

	result := StreamResult{}
	outputChunks := 0

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
			outputChunks++

			if usage := extractUsageFromChunk(chunk.Data); usage != nil {
				result.InputTokens = usage.PromptTokens
				result.OutputTokens = usage.CompletionTokens
			}
		}
	}

	if result.OutputTokens == 0 {
		result.OutputTokens = estimateTokens(outputChunks)
		result.Estimated = true
	}

	return result
}

func extractUsageFromChunk(data []byte) *adaptor.Usage {
	line := string(data)
	if !strings.HasPrefix(line, "data: ") {
		return nil
	}
	jsonStr := strings.TrimPrefix(line, "data: ")
	jsonStr = strings.TrimSpace(jsonStr)

	var chunk struct {
		Usage *adaptor.Usage `json:"usage"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &chunk); err != nil {
		return nil
	}
	if chunk.Usage != nil && chunk.Usage.TotalTokens > 0 {
		return chunk.Usage
	}
	return nil
}

func estimateTokens(chunks int) int {
	return chunks * 2
}
