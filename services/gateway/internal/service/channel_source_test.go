package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KeKe-Li/ai-token/services/gateway/internal/model"
)

func TestBuildRelayChannelDecryptsDatabaseAPIKey(t *testing.T) {
	const key = "unit-test-encryption-key"
	enc, err := Encrypt("sk-real-secret", key)
	if err != nil {
		t.Fatalf("加密测试密钥失败: %v", err)
	}

	ch, err := BuildRelayChannel(model.Channel{
		ID:        7,
		Name:      "OpenAI 主渠道",
		Provider:  "openai",
		BaseURL:   "https://api.example.com",
		APIKeyEnc: enc,
		Models:    []string{"gpt-4o"},
		Status:    1,
		Priority:  9,
		Weight:    3,
	}, key)
	if err != nil {
		t.Fatalf("转换数据库渠道失败: %v", err)
	}

	if ch.APIKey != "sk-real-secret" {
		t.Fatalf("期望解密后的 API Key 进入 relay channel，实际=%q", ch.APIKey)
	}
	if ch.ID != 7 || ch.Provider != "openai" || ch.Priority != 9 || ch.Weight != 3 {
		t.Fatalf("渠道核心路由字段未正确保留: %+v", ch)
	}
}

func TestBuildRelayChannelRejectsUndecryptableAPIKey(t *testing.T) {
	_, err := BuildRelayChannel(model.Channel{
		ID:        9,
		Name:      "损坏渠道",
		Provider:  "openai",
		BaseURL:   "https://api.example.com",
		APIKeyEnc: "not-base64",
		Models:    []string{"gpt-4o"},
		Status:    1,
		Priority:  1,
		Weight:    1,
	}, "unit-test-encryption-key")
	if err == nil {
		t.Fatal("无法解密的数据库 API Key 不应被静默当作明文继续使用")
	}
}

func TestProbeChannelTreatsCredentialErrorsAsFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/models" {
			t.Fatalf("探测应访问 OpenAI models 端点，实际路径=%s", r.URL.Path)
		}
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"bad key"}`))
	}))
	defer server.Close()

	ok, status, msg := ProbeChannel("openai", server.URL, "sk-bad")
	if ok {
		t.Fatalf("上游返回 401 时不应判定渠道可用，status=%d msg=%s", status, msg)
	}
	if status != http.StatusUnauthorized {
		t.Fatalf("应保留上游状态码，实际=%d", status)
	}
}
