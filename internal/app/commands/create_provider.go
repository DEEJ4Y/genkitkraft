package commands

import (
	"context"

	"github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/provider"
	"github.com/DEEJ4Y/genkitkraft/internal/ports/encryptor"
	providerrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/provider_repo"
)

type CreateProviderParams struct {
	Name         string
	ProviderType provider.ProviderType
	APIKey       string
	BaseURL      string
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

	if params.APIKey == "" {
		return CreateProviderResult{}, errors.NewAppError(errors.InvalidInput, "api key is required")
	}

	encryptedKey, err := c.enc.Encrypt(params.APIKey)
	if err != nil {
		return CreateProviderResult{}, errors.NewAppErrorf(errors.Internal, "encrypting api key: %v", err)
	}

	p := &provider.Provider{
		Name:         params.Name,
		ProviderType: params.ProviderType,
		APIKey:       encryptedKey,
		BaseURL:      params.BaseURL,
		Enabled:      true,
	}

	if err := c.repo.Create(ctx, p); err != nil {
		return CreateProviderResult{}, err
	}

	// Restore plaintext for response masking
	p.APIKey = params.APIKey

	return CreateProviderResult{Provider: p}, nil
}
