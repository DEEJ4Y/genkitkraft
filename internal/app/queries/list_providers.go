package queries

import (
	"context"

	"github.com/DEEJ4Y/genkitkraft/internal/domain/provider"
	providerrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/provider_repo"
)

type ListProvidersParams struct{}

type ListProvidersResult struct {
	Providers []*provider.Provider
}

type ListProvidersQuery struct {
	repo providerrepo.ProviderRepository
}

func NewListProvidersQuery(repo providerrepo.ProviderRepository) *ListProvidersQuery {
	return &ListProvidersQuery{repo: repo}
}

func (q *ListProvidersQuery) Execute(ctx context.Context, params ListProvidersParams) (ListProvidersResult, error) {
	providers, err := q.repo.List(ctx)
	if err != nil {
		return ListProvidersResult{}, err
	}
	return ListProvidersResult{Providers: providers}, nil
}
