package provider

import (
	"encoding/json"
	"time"
)

// ProviderType identifies a supported LLM provider.
type ProviderType string

const (
	GoogleAI    ProviderType = "google_ai"
	VertexAI    ProviderType = "vertex_ai"
	OpenAI      ProviderType = "openai"
	Anthropic   ProviderType = "anthropic"
	Ollama      ProviderType = "ollama"
	XAI         ProviderType = "xai"
	DeepSeek    ProviderType = "deepseek"
	AzureOpenAI    ProviderType = "azure_openai"
	Bedrock        ProviderType = "bedrock"
	AzureAIFoundry    ProviderType = "azure_ai_foundry"
	OpenAICompatible  ProviderType = "openai_compatible"
)

// Valid returns true if the provider type is a known value.
func (pt ProviderType) Valid() bool {
	switch pt {
	case GoogleAI, VertexAI, OpenAI, Anthropic, Ollama, XAI, DeepSeek, AzureOpenAI, Bedrock, AzureAIFoundry, OpenAICompatible:
		return true
	}
	return false
}

// Provider represents a configured LLM provider with API credentials.
type Provider struct {
	ID           string
	Name         string
	ProviderType ProviderType
	APIKey       *string         // Nullable; some providers don't use API keys
	BaseURL      string
	Config       ProviderConfig  // Parsed config (hydrated from RawConfig)
	RawConfig    json.RawMessage // Raw JSON for DB serialization
	Enabled      bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// MaskedAPIKey returns the API key with only the last 4 characters visible.
// Returns nil if the provider has no API key.
func (p *Provider) MaskedAPIKey() *string {
	if p.APIKey == nil {
		return nil
	}
	key := *p.APIKey
	var masked string
	if len(key) <= 4 {
		masked = "****"
	} else {
		masked = "****" + key[len(key)-4:]
	}
	return &masked
}
