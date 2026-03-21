package queries

import (
	"context"

	"github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/provider"
	"github.com/DEEJ4Y/genkitkraft/internal/ports/encryptor"
	providerrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/provider_repo"
)

type ListProvidersParams struct{}

type ListProvidersResult struct {
	Providers []*provider.Provider
}

type ListProvidersQuery struct {
	repo providerrepo.ProviderRepository
	enc  encryptor.Encryptor
}

func NewListProvidersQuery(repo providerrepo.ProviderRepository, enc encryptor.Encryptor) *ListProvidersQuery {
	return &ListProvidersQuery{repo: repo, enc: enc}
}

func (q *ListProvidersQuery) Execute(ctx context.Context, params ListProvidersParams) (ListProvidersResult, error) {
	providers, err := q.repo.List(ctx)
	if err != nil {
		return ListProvidersResult{}, err
	}

	for _, p := range providers {
		if p.APIKey != nil {
			decrypted, err := q.enc.Decrypt(*p.APIKey)
			if err != nil {
				return ListProvidersResult{}, errors.NewAppErrorf(errors.Internal, "decrypting api key: %v", err)
			}
			p.APIKey = &decrypted
		}
	}

	return ListProvidersResult{Providers: providers}, nil
}
