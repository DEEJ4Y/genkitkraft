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

// HttpToolDefinition contains everything needed to register and execute an HTTP tool.
type HttpToolDefinition struct {
	ID           string
	Name         string
	Description  string
	Method       string
	URL          string
	Headers      map[string]string
	BodyTemplate string
	InputSchema  json.RawMessage
}

// McpServerConfig contains connection info and tool selection for an MCP server.
type McpServerConfig struct {
	ServerID  string
	Transport string // "sse" or "streamableHttp"
	URL       string
	Headers   map[string]string
	SelectAll bool
	ToolNames []string // specific tools when SelectAll is false
}

// ChatRequest contains all information needed to make a chat completion request.
type ChatRequest struct {
	ProviderType       string
	APIKey             string
	BaseURL            string
	Config             json.RawMessage // Provider-specific config
	ModelID            string
	SystemPrompt       string
	Messages           []ChatMessage
	TemperatureEnabled bool
	Temperature        float64
	TopPEnabled        bool
	TopP               float64
	TopKEnabled        bool
	TopK               int
	MaxToolCalls       int
	HttpTools          []HttpToolDefinition
	McpServers         []McpServerConfig
}

// ChatProvider defines the contract for chat completions from LLM providers.
type ChatProvider interface {
	// ChatStream sends messages and streams the response token by token via the channel.
	// The string channel receives content deltas. It is closed when the response is complete.
	// The error channel receives at most one error, then is closed.
	ChatStream(ctx context.Context, req ChatRequest) (<-chan string, <-chan error)

	// Chat sends messages and returns the full response as a single string.
	// Use this for providers that do not support streaming.
	Chat(ctx context.Context, req ChatRequest) (string, error)
}
