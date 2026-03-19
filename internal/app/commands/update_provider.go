package commands

import (
	"context"

	"github.com/DEEJ4Y/genkitkraft/internal/domain/provider"
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
}

func NewUpdateProviderCommand(repo providerrepo.ProviderRepository) *UpdateProviderCommand {
	return &UpdateProviderCommand{repo: repo}
}

func (c *UpdateProviderCommand) Execute(ctx context.Context, params UpdateProviderParams) (UpdateProviderResult, error) {
	p, err := c.repo.GetByID(ctx, params.ID)
	if err != nil {
		return UpdateProviderResult{}, err
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

	if err := c.repo.Update(ctx, p); err != nil {
		return UpdateProviderResult{}, err
	}

	return UpdateProviderResult{Provider: p}, nil
}
