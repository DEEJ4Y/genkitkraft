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
	p.APIKey, err = c.enc.Decrypt(p.APIKey)
	if err != nil {
		return UpdateProviderResult{}, errors.NewAppErrorf(errors.Internal, "decrypting api key: %v", err)
	}

	if params.Name != nil {
		p.Name = *params.Name
	}
	if params.APIKey != nil && *params.APIKey != "" {
		p.APIKey = *params.APIKey
	}
	if params.BaseURL != nil {
		p.BaseURL = *params.BaseURL
	}
	if params.Enabled != nil {
		p.Enabled = *params.Enabled
	}

	// Save plaintext for response
	plaintextKey := p.APIKey

	// Encrypt for storage
	p.APIKey, err = c.enc.Encrypt(p.APIKey)
	if err != nil {
		return UpdateProviderResult{}, errors.NewAppErrorf(errors.Internal, "encrypting api key: %v", err)
	}

	if err := c.repo.Update(ctx, p); err != nil {
		return UpdateProviderResult{}, err
	}

	// Restore plaintext for response masking
	p.APIKey = plaintextKey

	return UpdateProviderResult{Provider: p}, nil
}
