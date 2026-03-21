package commands

import (
	"context"
	"encoding/json"

	"github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/provider"
	"github.com/DEEJ4Y/genkitkraft/internal/ports/encryptor"
	providerrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/provider_repo"
)

type CreateProviderParams struct {
	Name         string
	ProviderType provider.ProviderType
	APIKey       *string
	BaseURL      string
	Config       map[string]string
}

type CreateProviderResult struct {
	Provider *provider.Provider
}

type CreateProviderCommand struct {
	repo providerrepo.ProviderRepository
	enc  encryptor.Encryptor
}

func NewCreateProviderCommand(repo providerrepo.ProviderRepository, enc encryptor.Encryptor) *CreateProviderCommand {
	return &CreateProviderCommand{repo: repo, enc: enc}
}

func (c *CreateProviderCommand) Execute(ctx context.Context, params CreateProviderParams) (CreateProviderResult, error) {
	if !params.ProviderType.Valid() {
		return CreateProviderResult{}, errors.NewAppErrorf(errors.InvalidInput, "invalid provider type: %s", params.ProviderType)
	}

	if params.Name == "" {
		return CreateProviderResult{}, errors.NewAppError(errors.InvalidInput, "name is required")
	}

	// Validate provider-specific requirements
	if err := validateProviderRequirements(params.ProviderType, params.APIKey, params.BaseURL, params.Config); err != nil {
		return CreateProviderResult{}, err
	}

	// Marshal config to JSON
	rawConfig, err := marshalConfig(params.Config)
	if err != nil {
		return CreateProviderResult{}, errors.NewAppErrorf(errors.Internal, "marshaling config: %v", err)
	}

	// Parse and validate config
	cfg, err := provider.ParseProviderConfig(string(params.ProviderType), rawConfig)
	if err != nil {
		return CreateProviderResult{}, errors.NewAppErrorf(errors.InvalidInput, "invalid config: %v", err)
	}
	if err := cfg.Validate(); err != nil {
		return CreateProviderResult{}, errors.NewAppErrorf(errors.InvalidInput, "config validation: %v", err)
	}

	// Encrypt API key if provided
	var encryptedKey *string
	if params.APIKey != nil && *params.APIKey != "" {
		encrypted, err := c.enc.Encrypt(*params.APIKey)
		if err != nil {
			return CreateProviderResult{}, errors.NewAppErrorf(errors.Internal, "encrypting api key: %v", err)
		}
		encryptedKey = &encrypted
	}

	p := &provider.Provider{
		Name:         params.Name,
		ProviderType: params.ProviderType,
		APIKey:       encryptedKey,
		BaseURL:      params.BaseURL,
		Config:       cfg,
		RawConfig:    rawConfig,
		Enabled:      true,
	}

	if err := c.repo.Create(ctx, p); err != nil {
		return CreateProviderResult{}, err
	}

	// Restore plaintext key for response masking
	p.APIKey = params.APIKey

	return CreateProviderResult{Provider: p}, nil
}

// validateProviderRequirements checks provider-specific field requirements.
func validateProviderRequirements(pt provider.ProviderType, apiKey *string, baseURL string, config map[string]string) error {
	hasAPIKey := apiKey != nil && *apiKey != ""

	switch pt {
	case provider.GoogleAI, provider.OpenAI, provider.Anthropic, provider.XAI, provider.DeepSeek:
		if !hasAPIKey {
			return errors.NewAppError(errors.InvalidInput, "api key is required")
		}
	case provider.Ollama:
		if baseURL == "" {
			return errors.NewAppError(errors.InvalidInput, "base_url is required for Ollama")
		}
	case provider.AzureOpenAI:
		if !hasAPIKey {
			return errors.NewAppError(errors.InvalidInput, "api key is required for Azure OpenAI")
		}
		if baseURL == "" {
			return errors.NewAppError(errors.InvalidInput, "base_url is required for Azure OpenAI")
		}
	case provider.VertexAI:
		// config.project and config.location validated by VertexAIConfig.Validate()
	case provider.Bedrock:
		// config fields validated by BedrockConfig.Validate()
	case provider.AzureAIFoundry:
		if !hasAPIKey {
			return errors.NewAppError(errors.InvalidInput, "api key is required for Azure AI Foundry")
		}
		if baseURL == "" {
			return errors.NewAppError(errors.InvalidInput, "base_url is required for Azure AI Foundry")
		}
	case provider.OpenAICompatible:
		if !hasAPIKey {
			return errors.NewAppError(errors.InvalidInput, "api key is required for OpenAI Compatible")
		}
		if baseURL == "" {
			return errors.NewAppError(errors.InvalidInput, "base_url is required for OpenAI Compatible")
		}
	}

	return nil
}

// marshalConfig converts a string map to json.RawMessage.
func marshalConfig(config map[string]string) (json.RawMessage, error) {
	if len(config) == 0 {
		return json.RawMessage("{}"), nil
	}
	data, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}
