package provider

import (
	"encoding/json"
	"fmt"
)

// ProviderConfig is the interface for provider-specific configuration.
type ProviderConfig interface {
	Validate() error
}

// GoogleAIConfig holds Google AI-specific configuration.
type GoogleAIConfig struct{}

func (c *GoogleAIConfig) Validate() error { return nil }

// VertexAIConfig holds Vertex AI-specific configuration.
type VertexAIConfig struct {
	Project  string `json:"project"`
	Location string `json:"location"`
}

func (c *VertexAIConfig) Validate() error {
	if c.Project == "" {
		return fmt.Errorf("project is required for Vertex AI")
	}
	if c.Location == "" {
		return fmt.Errorf("location is required for Vertex AI")
	}
	return nil
}

// OpenAIConfig holds OpenAI-specific configuration.
type OpenAIConfig struct{}

func (c *OpenAIConfig) Validate() error { return nil }

// AnthropicConfig holds Anthropic-specific configuration.
type AnthropicConfig struct{}

func (c *AnthropicConfig) Validate() error { return nil }

// OllamaConfig holds Ollama-specific configuration.
type OllamaConfig struct{}

func (c *OllamaConfig) Validate() error { return nil }

// XAIConfig holds xAI-specific configuration.
type XAIConfig struct{}

func (c *XAIConfig) Validate() error { return nil }

// DeepSeekConfig holds DeepSeek-specific configuration.
type DeepSeekConfig struct{}

func (c *DeepSeekConfig) Validate() error { return nil }

// AzureOpenAIConfig holds Azure OpenAI-specific configuration.
type AzureOpenAIConfig struct {
	DeploymentName string `json:"deployment_name"`
	APIVersion     string `json:"api_version"`
}

func (c *AzureOpenAIConfig) Validate() error {
	if c.DeploymentName == "" {
		return fmt.Errorf("deployment_name is required for Azure OpenAI")
	}
	if c.APIVersion == "" {
		return fmt.Errorf("api_version is required for Azure OpenAI")
	}
	return nil
}

// BedrockConfig holds AWS Bedrock-specific configuration.
type BedrockConfig struct {
	Region          string `json:"region"`
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey  string `json:"secret_access_key"`
	SessionToken    string `json:"session_token,omitempty"`
}

func (c *BedrockConfig) Validate() error {
	if c.Region == "" {
		return fmt.Errorf("region is required for Bedrock")
	}
	if c.AccessKeyID == "" {
		return fmt.Errorf("access_key_id is required for Bedrock")
	}
	if c.SecretAccessKey == "" {
		return fmt.Errorf("secret_access_key is required for Bedrock")
	}
	return nil
}

// ParseProviderConfig unmarshals raw JSON into the correct config struct for the provider type.
func ParseProviderConfig(providerType string, raw json.RawMessage) (ProviderConfig, error) {
	if len(raw) == 0 || string(raw) == "{}" || string(raw) == "null" {
		return defaultConfig(providerType)
	}

	var cfg ProviderConfig
	switch ProviderType(providerType) {
	case GoogleAI:
		cfg = &GoogleAIConfig{}
	case VertexAI:
		cfg = &VertexAIConfig{}
	case OpenAI:
		cfg = &OpenAIConfig{}
	case Anthropic:
		cfg = &AnthropicConfig{}
	case Ollama:
		cfg = &OllamaConfig{}
	case XAI:
		cfg = &XAIConfig{}
	case DeepSeek:
		cfg = &DeepSeekConfig{}
	case AzureOpenAI:
		cfg = &AzureOpenAIConfig{}
	case Bedrock:
		cfg = &BedrockConfig{}
	default:
		return nil, fmt.Errorf("unknown provider type: %s", providerType)
	}

	if err := json.Unmarshal(raw, cfg); err != nil {
		return nil, fmt.Errorf("parsing config for %s: %w", providerType, err)
	}
	return cfg, nil
}

func defaultConfig(providerType string) (ProviderConfig, error) {
	switch ProviderType(providerType) {
	case GoogleAI:
		return &GoogleAIConfig{}, nil
	case VertexAI:
		return &VertexAIConfig{}, nil
	case OpenAI:
		return &OpenAIConfig{}, nil
	case Anthropic:
		return &AnthropicConfig{}, nil
	case Ollama:
		return &OllamaConfig{}, nil
	case XAI:
		return &XAIConfig{}, nil
	case DeepSeek:
		return &DeepSeekConfig{}, nil
	case AzureOpenAI:
		return &AzureOpenAIConfig{}, nil
	case Bedrock:
		return &BedrockConfig{}, nil
	default:
		return nil, fmt.Errorf("unknown provider type: %s", providerType)
	}
}
