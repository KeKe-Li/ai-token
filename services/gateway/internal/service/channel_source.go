package service

import (
	"fmt"

	"github.com/KeKe-Li/ai-token/services/gateway/internal/model"
	"github.com/KeKe-Li/ai-token/services/gateway/internal/relay"
)

// BuildRelayChannel 将数据库中的渠道配置转换为运行时 relay 渠道。
// 注意：数据库中的供应商密钥必须是加密值，不能在解密失败时降级为明文使用。
func BuildRelayChannel(ch model.Channel, encryptionKey string) (relay.Channel, error) {
	if ch.APIKeyEnc == "" {
		return relay.Channel{}, fmt.Errorf("channel %d api key is empty", ch.ID)
	}
	if encryptionKey == "" {
		return relay.Channel{}, fmt.Errorf("channel %d encryption key is empty", ch.ID)
	}

	apiKey, err := Decrypt(ch.APIKeyEnc, encryptionKey)
	if err != nil {
		return relay.Channel{}, fmt.Errorf("decrypt channel %d api key: %w", ch.ID, err)
	}

	return relay.Channel{
		ID:            ch.ID,
		Name:          ch.Name,
		Provider:      ch.Provider,
		BaseURL:       ch.BaseURL,
		APIKey:        apiKey,
		Models:        ch.Models,
		Status:        ch.Status,
		Priority:      ch.Priority,
		Weight:        ch.Weight,
		CooldownUntil: ch.CooldownUntil,
	}, nil
}
