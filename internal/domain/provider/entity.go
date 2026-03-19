package provider

import "time"

// ProviderType identifies a supported LLM provider.
type ProviderType string

const (
	Anthropic ProviderType = "anthropic"
	OpenAI    ProviderType = "openai"
	GoogleAI  ProviderType = "googleai"
)

// Valid returns true if the provider type is a known value.
func (pt ProviderType) Valid() bool {
	switch pt {
	case Anthropic, OpenAI, GoogleAI:
		return true
	}
	return false
}

// Provider represents a configured LLM provider with API credentials.
type Provider struct {
	ID           string
	Name         string
	ProviderType ProviderType
	APIKey       string
	BaseURL      string
	Enabled      bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// MaskedAPIKey returns the API key with only the last 4 characters visible.
func (p *Provider) MaskedAPIKey() string {
	if len(p.APIKey) <= 4 {
		return "****"
	}
	return "****" + p.APIKey[len(p.APIKey)-4:]
}
