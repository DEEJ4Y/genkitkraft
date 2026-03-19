package commands

import (
	"context"

	"github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/provider"
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
}

func NewCreateProviderCommand(repo providerrepo.ProviderRepository) *CreateProviderCommand {
	return &CreateProviderCommand{repo: repo}
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

	p := &provider.Provider{
		Name:         params.Name,
		ProviderType: params.ProviderType,
		APIKey:       params.APIKey,
		BaseURL:      params.BaseURL,
		Enabled:      true,
	}

	if err := c.repo.Create(ctx, p); err != nil {
		return CreateProviderResult{}, err
	}

	return CreateProviderResult{Provider: p}, nil
}
