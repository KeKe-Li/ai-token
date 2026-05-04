package service

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthChecker struct {
	db       *pgxpool.Pool
	interval time.Duration
	stop     chan struct{}
}

func NewHealthChecker(db *pgxpool.Pool, interval time.Duration) *HealthChecker {
	return &HealthChecker{
		db:       db,
		interval: interval,
		stop:     make(chan struct{}),
	}
}

func (h *HealthChecker) Start() {
	go func() {
		ticker := time.NewTicker(h.interval)
		defer ticker.Stop()

		h.checkAll()

		for {
			select {
			case <-ticker.C:
				h.checkAll()
			case <-h.stop:
				return
			}
		}
	}()
	log.Printf("渠道健康检测已启动，间隔 %v", h.interval)
}

func (h *HealthChecker) Stop() {
	close(h.stop)
}

func (h *HealthChecker) checkAll() {
	ctx := context.Background()

	rows, err := h.db.Query(ctx,
		`SELECT id, name, provider, base_url, status FROM channels WHERE status != 2`)
	if err != nil {
		log.Printf("健康检测: 查询渠道失败: %v", err)
		return
	}
	defer rows.Close()

	type channel struct {
		id       int64
		name     string
		provider string
		baseURL  string
		status   int
	}

	var channels []channel
	for rows.Next() {
		var ch channel
		if err := rows.Scan(&ch.id, &ch.name, &ch.provider, &ch.baseURL, &ch.status); err != nil {
			continue
		}
		channels = append(channels, ch)
	}

	for _, ch := range channels {
		healthy := h.probe(ch.provider, ch.baseURL)
		if healthy && ch.status == 3 {
			h.db.Exec(ctx, `UPDATE channels SET status = 1, cooldown_until = NULL, updated_at = NOW() WHERE id = $1`, ch.id)
			log.Printf("渠道 %s (#%d) 恢复正常", ch.name, ch.id)
		} else if !healthy && ch.status == 1 {
			cooldown := time.Now().Add(5 * time.Minute)
			h.db.Exec(ctx, `UPDATE channels SET status = 3, cooldown_until = $1, updated_at = NOW() WHERE id = $2`, cooldown, ch.id)
			log.Printf("渠道 %s (#%d) 探测失败，进入冷却", ch.name, ch.id)
		}
	}
}

func (h *HealthChecker) probe(provider, baseURL string) bool {
	url := strings.TrimRight(baseURL, "/")

	switch provider {
	case "openai", "deepseek":
		url += "/v1/models"
	case "anthropic":
		url += "/v1/messages"
	case "google":
		url += "/v1beta/models"
	default:
		url += "/health"
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if provider == "anthropic" {
		return resp.StatusCode == 401 || resp.StatusCode == 200
	}

	return resp.StatusCode < 500
}

func ProbeChannel(provider, baseURL, apiKey string) (bool, int, string) {
	ok, statusCode, _, message := ProbeChannelDetailed(provider, baseURL, apiKey)
	return ok, statusCode, message
}

func ProbeChannelDetailed(provider, baseURL, apiKey string) (bool, int, int, string) {
	url := strings.TrimRight(baseURL, "/")

	switch provider {
	case "openai", "deepseek":
		url += "/v1/models"
	case "anthropic":
		url += "/v1/messages"
	case "google":
		url += "/v1beta/models"
	default:
		url += "/health"
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return false, 0, 0, err.Error()
	}

	if apiKey != "" {
		switch provider {
		case "anthropic":
			req.Header.Set("x-api-key", apiKey)
			req.Header.Set("anthropic-version", "2023-06-01")
		default:
			req.Header.Set("Authorization", "Bearer "+apiKey)
		}
	}

	start := time.Now()
	resp, err := client.Do(req)
	latency := int(time.Since(start).Milliseconds())
	if err != nil {
		return false, 0, latency, fmt.Sprintf("连接失败: %v", err)
	}
	defer resp.Body.Close()

	if apiKey != "" && (resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden) {
		return false, resp.StatusCode, latency, fmt.Sprintf("认证失败 %d", resp.StatusCode)
	}

	if resp.StatusCode >= 500 {
		return false, resp.StatusCode, latency, fmt.Sprintf("服务端错误 %d", resp.StatusCode)
	}

	return true, resp.StatusCode, latency, fmt.Sprintf("状态 %d, 延迟 %dms", resp.StatusCode, latency)
}
