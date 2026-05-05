package relay

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/KeKe-Li/ai-token/services/gateway/internal/relay/adaptor"
)

type UsageRecord struct {
	UserID          int64
	APIKeyID        int64
	ChannelID       int64
	Model           string
	Method          string
	Path            string
	StatusCode      int
	InputTokens     int
	OutputTokens    int
	Cost            int64
	ReservedAmount  int64
	WalletHoldID    int64
	LatencyMs       int
	Error           string
	IP              string
	EstimatedTokens bool
	BillingStatus   string
	BillingNote     string
}

type LogWriter interface {
	WriteLog(ctx context.Context, record *UsageRecord) error
}

type CooldownStore interface {
	SetCooldown(ctx context.Context, channelID int64, duration time.Duration) error
	IsCoolingDown(ctx context.Context, channelID int64) bool
}

type RelayEngine struct {
	router        *Router
	registry      *adaptor.Registry
	usageCh       chan UsageRecord
	logWriter     LogWriter
	cooldownStore CooldownStore
}

func NewRelayEngine(logWriter LogWriter, cooldownStore CooldownStore) *RelayEngine {
	registry := adaptor.NewRegistry()
	registry.Register("openai", adaptor.NewOpenAIAdaptor())
	registry.Register("anthropic", adaptor.NewAnthropicAdaptor())
	registry.Register("google", adaptor.NewGoogleAdaptor())
	registry.Register("deepseek", adaptor.NewDeepSeekAdaptor())

	engine := &RelayEngine{
		router:        NewRouter(cooldownStore),
		registry:      registry,
		usageCh:       make(chan UsageRecord, 1024),
		logWriter:     logWriter,
		cooldownStore: cooldownStore,
	}

	go engine.processUsageRecords()

	return engine
}

func (e *RelayEngine) ChatCompletion(ctx context.Context, channels []Channel, req *adaptor.ChatRequest, maxRetries int) (*adaptor.ChatResponse, *Channel, error) {
	var lastErr error
	triedChannels := make(map[int64]bool)

	for attempt := 0; attempt <= maxRetries; attempt++ {
		available := filterUntried(channels, triedChannels)
		channel, err := e.router.SelectChannel(available, req.Model)
		if err != nil {
			if lastErr != nil {
				return nil, nil, fmt.Errorf("all channels exhausted, last error: %w", lastErr)
			}
			return nil, nil, err
		}

		triedChannels[channel.ID] = true

		a, ok := e.registry.Get(channel.Provider)
		if !ok {
			lastErr = fmt.Errorf("no adaptor for provider %s", channel.Provider)
			continue
		}

		cfg := &adaptor.ChannelConfig{
			Provider: channel.Provider,
			BaseURL:  channel.BaseURL,
			APIKey:   channel.APIKey,
			Model:    req.Model,
		}

		resp, err := a.ChatCompletion(ctx, cfg, req)
		if err != nil {
			lastErr = err
			e.cooldownChannel(channel.ID)
			log.Printf("渠道 %s (#%d) 失败, 尝试下一个: %v", channel.Name, channel.ID, err)
			continue
		}

		return resp, channel, nil
	}

	return nil, nil, fmt.Errorf("all retries exhausted: %w", lastErr)
}

func (e *RelayEngine) ChatCompletionStream(ctx context.Context, channels []Channel, req *adaptor.ChatRequest, maxRetries int) (<-chan adaptor.StreamChunk, *Channel, error) {
	var lastErr error
	triedChannels := make(map[int64]bool)

	for attempt := 0; attempt <= maxRetries; attempt++ {
		available := filterUntried(channels, triedChannels)
		channel, err := e.router.SelectChannel(available, req.Model)
		if err != nil {
			if lastErr != nil {
				return nil, nil, fmt.Errorf("all channels exhausted, last error: %w", lastErr)
			}
			return nil, nil, err
		}

		triedChannels[channel.ID] = true

		a, ok := e.registry.Get(channel.Provider)
		if !ok {
			lastErr = fmt.Errorf("no adaptor for provider %s", channel.Provider)
			continue
		}

		cfg := &adaptor.ChannelConfig{
			Provider: channel.Provider,
			BaseURL:  channel.BaseURL,
			APIKey:   channel.APIKey,
			Model:    req.Model,
		}

		stream, err := a.ChatCompletionStream(ctx, cfg, req)
		if err != nil {
			lastErr = err
			e.cooldownChannel(channel.ID)
			log.Printf("渠道 %s (#%d) 流式失败, 尝试下一个: %v", channel.Name, channel.ID, err)
			continue
		}

		return stream, channel, nil
	}

	return nil, nil, fmt.Errorf("all retries exhausted: %w", lastErr)
}

func (e *RelayEngine) RecordUsage(record UsageRecord) {
	select {
	case e.usageCh <- record:
	default:
		log.Printf("用量记录通道已满，丢弃记录")
	}
}

func (e *RelayEngine) processUsageRecords() {
	for record := range e.usageCh {
		if e.logWriter != nil {
			if err := e.logWriter.WriteLog(context.Background(), &record); err != nil {
				log.Printf("写入用量记录失败: %v", err)
			}
		}
	}
}

func (e *RelayEngine) cooldownChannel(channelID int64) {
	if e.cooldownStore != nil {
		if err := e.cooldownStore.SetCooldown(context.Background(), channelID, 5*time.Minute); err != nil {
			log.Printf("设置渠道冷却失败: %v", err)
		}
	}
	log.Printf("渠道 #%d 进入冷却期 5 分钟", channelID)
}

func filterUntried(channels []Channel, tried map[int64]bool) []Channel {
	var result []Channel
	for _, ch := range channels {
		if !tried[ch.ID] {
			result = append(result, ch)
		}
	}
	return result
}
