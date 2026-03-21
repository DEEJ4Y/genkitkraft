package commands

import (
	"context"

	"github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/provider"
	"github.com/DEEJ4Y/genkitkraft/internal/ports/encryptor"
	providerrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/provider_repo"
)

type UpdateProviderParams struct {
	ID      string
	Name    *string
	APIKey  *string
	BaseURL *string
	Config  *map[string]string
	Enabled *bool
}

type UpdateProviderResult struct {
	Provider *provider.Provider
}

type UpdateProviderCommand struct {
	repo providerrepo.ProviderRepository
	enc  encryptor.Encryptor
}

func NewUpdateProviderCommand(repo providerrepo.ProviderRepository, enc encryptor.Encryptor) *UpdateProviderCommand {
	return &UpdateProviderCommand{repo: repo, enc: enc}
}

func (c *UpdateProviderCommand) Execute(ctx context.Context, params UpdateProviderParams) (UpdateProviderResult, error) {
	p, err := c.repo.GetByID(ctx, params.ID)
	if err != nil {
		return UpdateProviderResult{}, err
	}

	// Decrypt existing key so we work with plaintext in memory
	if p.APIKey != nil {
		decrypted, err := c.enc.Decrypt(*p.APIKey)
		if err != nil {
			return UpdateProviderResult{}, errors.NewAppErrorf(errors.Internal, "decrypting api key: %v", err)
		}
		p.APIKey = &decrypted
	}

	if params.Name != nil {
		p.Name = *params.Name
	}
	if params.APIKey != nil && *params.APIKey != "" {
		p.APIKey = params.APIKey
	}
	if params.BaseURL != nil {
		p.BaseURL = *params.BaseURL
	}
	if params.Enabled != nil {
		p.Enabled = *params.Enabled
	}

	// Update config if provided
	if params.Config != nil {
		rawConfig, err := marshalConfig(*params.Config)
		if err != nil {
			return UpdateProviderResult{}, errors.NewAppErrorf(errors.Internal, "marshaling config: %v", err)
		}
		cfg, err := provider.ParseProviderConfig(string(p.ProviderType), rawConfig)
		if err != nil {
			return UpdateProviderResult{}, errors.NewAppErrorf(errors.InvalidInput, "invalid config: %v", err)
		}
		if err := cfg.Validate(); err != nil {
			return UpdateProviderResult{}, errors.NewAppErrorf(errors.InvalidInput, "config validation: %v", err)
		}
		p.Config = cfg
		p.RawConfig = rawConfig
	}

	// Save plaintext for response
	plaintextKey := p.APIKey

	// Encrypt for storage
	if p.APIKey != nil {
		encrypted, err := c.enc.Encrypt(*p.APIKey)
		if err != nil {
			return UpdateProviderResult{}, errors.NewAppErrorf(errors.Internal, "encrypting api key: %v", err)
		}
		p.APIKey = &encrypted
	}

	if err := c.repo.Update(ctx, p); err != nil {
		return UpdateProviderResult{}, err
	}

	// Restore plaintext for response masking
	p.APIKey = plaintextKey

	return UpdateProviderResult{Provider: p}, nil
}
