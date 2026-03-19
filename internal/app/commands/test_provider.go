package commands

import (
	"context"

	"github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	providerrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/provider_repo"
	providertester "github.com/DEEJ4Y/genkitkraft/internal/ports/provider_tester"
)

type TestProviderParams struct {
	ID string
}

type TestProviderResult struct {
	Success bool
	Message string
}

type TestProviderCommand struct {
	repo   providerrepo.ProviderRepository
	tester providertester.Tester
}

func NewTestProviderCommand(repo providerrepo.ProviderRepository, tester providertester.Tester) *TestProviderCommand {
	return &TestProviderCommand{repo: repo, tester: tester}
}

func (c *TestProviderCommand) Execute(ctx context.Context, params TestProviderParams) (TestProviderResult, error) {
	p, err := c.repo.GetByID(ctx, params.ID)
	if err != nil {
		return TestProviderResult{}, err
	}

	success, message, err := c.tester.Test(ctx, p)
	if err != nil {
		return TestProviderResult{}, errors.NewAppErrorf(errors.Internal, "testing provider: %v", err)
	}

	return TestProviderResult{Success: success, Message: message}, nil
}
