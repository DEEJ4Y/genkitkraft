package commands

import (
	"context"

	"github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	"github.com/DEEJ4Y/genkitkraft/internal/ports/encryptor"
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
	enc    encryptor.Encryptor
}

func NewTestProviderCommand(repo providerrepo.ProviderRepository, tester providertester.Tester, enc encryptor.Encryptor) *TestProviderCommand {
	return &TestProviderCommand{repo: repo, tester: tester, enc: enc}
}

func (c *TestProviderCommand) Execute(ctx context.Context, params TestProviderParams) (TestProviderResult, error) {
	p, err := c.repo.GetByID(ctx, params.ID)
	if err != nil {
		return TestProviderResult{}, err
	}

	// Decrypt API key so the tester can use it in HTTP headers
	if p.APIKey != nil {
		decrypted, err := c.enc.Decrypt(*p.APIKey)
		if err != nil {
			return TestProviderResult{}, errors.NewAppErrorf(errors.Internal, "decrypting api key: %v", err)
		}
		p.APIKey = &decrypted
	}

	success, message, err := c.tester.Test(ctx, p)
	if err != nil {
		return TestProviderResult{}, errors.NewAppErrorf(errors.Internal, "testing provider: %v", err)
	}

	return TestProviderResult{Success: success, Message: message}, nil
}
