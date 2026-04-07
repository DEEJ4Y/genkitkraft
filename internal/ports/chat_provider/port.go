package chatprovider

import (
	"context"
	"encoding/json"
)

// ChatMessage represents a single message in a conversation.
type ChatMessage struct {
	Role    string // "user", "assistant", or "system"
	Content string
}

// ChatRequest contains all information needed to make a chat completion request.
type ChatRequest struct {
	ProviderType string
	APIKey       string
	BaseURL      string
	Config       json.RawMessage // Provider-specific config
	ModelID      string
	SystemPrompt string
	Messages     []ChatMessage
	Temperature  float64
	TopP         float64
	TopK         int
}

// ChatProvider defines the contract for streaming chat completions from LLM providers.
type ChatProvider interface {
	// ChatStream sends messages and streams the response token by token via the channel.
	// The string channel receives content deltas. It is closed when the response is complete.
	// The error channel receives at most one error, then is closed.
	ChatStream(ctx context.Context, req ChatRequest) (<-chan string, <-chan error)
}
