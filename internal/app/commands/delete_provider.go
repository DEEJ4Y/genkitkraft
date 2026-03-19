package commands

import (
	"context"

	providerrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/provider_repo"
)

type DeleteProviderParams struct {
	ID string
}

type DeleteProviderCommand struct {
	repo providerrepo.ProviderRepository
}

func NewDeleteProviderCommand(repo providerrepo.ProviderRepository) *DeleteProviderCommand {
	return &DeleteProviderCommand{repo: repo}
}

func (c *DeleteProviderCommand) Execute(ctx context.Context, params DeleteProviderParams) error {
	return c.repo.Delete(ctx, params.ID)
}
