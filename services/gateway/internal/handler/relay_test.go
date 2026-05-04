package handler

import (
	"encoding/json"
	"testing"
	"time"
)

func TestModelAllowedAllowsUnrestrictedKeys(t *testing.T) {
	if !modelAllowed("gpt-4o", nil) {
		t.Fatal("nil 模型列表应表示不限制模型")
	}
	if !modelAllowed("gpt-4o", []string{}) {
		t.Fatal("空模型列表应表示不限制模型")
	}
}

func TestModelAllowedRejectsUnlistedModel(t *testing.T) {
	if modelAllowed("gpt-4o", []string{"claude-sonnet-4-6"}) {
		t.Fatal("API Key 设置模型白名单后，不应允许未列出的模型")
	}
}

func TestModelAllowedSupportsWildcardPrefix(t *testing.T) {
	if !modelAllowed("gpt-4o-mini", []string{"gpt-*"}) {
		t.Fatal("模型白名单应支持前缀通配符")
	}
	if modelAllowed("claude-sonnet-4-6", []string{"gpt-*"}) {
		t.Fatal("前缀通配符不应匹配其他供应商模型")
	}
}

func TestCreateKeyRequestAcceptsExpiresAt(t *testing.T) {
	body := []byte(`{"name":"ci-key","models":["gpt-4o"],"expires_at":"2030-01-02T03:04:05Z"}`)
	var req createKeyRequest
	if err := json.Unmarshal(body, &req); err != nil {
		t.Fatalf("解析创建密钥请求失败: %v", err)
	}
	if req.ExpiresAt == nil {
		t.Fatal("expires_at 应被后端请求结构接收并传入存储层")
	}
	if got := req.ExpiresAt.UTC().Format(time.RFC3339); got != "2030-01-02T03:04:05Z" {
		t.Fatalf("expires_at 解析不正确: %s", got)
	}
}
