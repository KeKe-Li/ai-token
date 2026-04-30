package adaptor

import (
	"context"
)

type DeepSeekAdaptor struct {
	openai *OpenAIAdaptor
}

func NewDeepSeekAdaptor() *DeepSeekAdaptor {
	return &DeepSeekAdaptor{
		openai: NewOpenAIAdaptor(),
	}
}

func (a *DeepSeekAdaptor) Name() string {
	return "deepseek"
}

func (a *DeepSeekAdaptor) ChatCompletion(ctx context.Context, cfg *ChannelConfig, req *ChatRequest) (*ChatResponse, error) {
	return a.openai.ChatCompletion(ctx, cfg, req)
}

func (a *DeepSeekAdaptor) ChatCompletionStream(ctx context.Context, cfg *ChannelConfig, req *ChatRequest) (<-chan StreamChunk, error) {
	return a.openai.ChatCompletionStream(ctx, cfg, req)
}

func (a *DeepSeekAdaptor) Models(_ context.Context, _ *ChannelConfig) ([]string, error) {
	return []string{
		"deepseek-chat",
		"deepseek-reasoner",
	}, nil
}
