package adaptor

import (
	"context"
)

type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []Message     `json:"messages"`
	Stream      bool          `json:"stream,omitempty"`
	Temperature *float64      `json:"temperature,omitempty"`
	MaxTokens   *int          `json:"max_tokens,omitempty"`
	TopP        *float64      `json:"top_p,omitempty"`
	Tools       []Tool        `json:"tools,omitempty"`
	Stop        []string      `json:"stop,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

type Function struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Parameters  any    `json:"parameters,omitempty"`
}

type ChatResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

type Choice struct {
	Index        int      `json:"index"`
	Message      *Message `json:"message,omitempty"`
	Delta        *Message `json:"delta,omitempty"`
	FinishReason *string  `json:"finish_reason,omitempty"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type StreamChunk struct {
	Data  []byte
	Error error
	Done  bool
}

type ChannelConfig struct {
	Provider string
	BaseURL  string
	APIKey   string
	Model    string
}

type Adaptor interface {
	Name() string
	ChatCompletion(ctx context.Context, cfg *ChannelConfig, req *ChatRequest) (*ChatResponse, error)
	ChatCompletionStream(ctx context.Context, cfg *ChannelConfig, req *ChatRequest) (<-chan StreamChunk, error)
	Models(ctx context.Context, cfg *ChannelConfig) ([]string, error)
}

type Registry struct {
	adaptors map[string]Adaptor
}

func NewRegistry() *Registry {
	return &Registry{
		adaptors: make(map[string]Adaptor),
	}
}

func (r *Registry) Register(name string, a Adaptor) {
	r.adaptors[name] = a
}

func (r *Registry) Get(name string) (Adaptor, bool) {
	a, ok := r.adaptors[name]
	return a, ok
}

func (r *Registry) List() []string {
	names := make([]string, 0, len(r.adaptors))
	for name := range r.adaptors {
		names = append(names, name)
	}
	return names
}

