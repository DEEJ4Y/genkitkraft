package queries

import (
	"context"

	"github.com/DEEJ4Y/genkitkraft/internal/domain/provider"
	providerrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/provider_repo"
)

type GetProviderParams struct {
	ID string
}

type GetProviderResult struct {
	Provider *provider.Provider
}

type GetProviderQuery struct {
	repo providerrepo.ProviderRepository
}

func NewGetProviderQuery(repo providerrepo.ProviderRepository) *GetProviderQuery {
	return &GetProviderQuery{repo: repo}
}

func (q *GetProviderQuery) Execute(ctx context.Context, params GetProviderParams) (GetProviderResult, error) {
	p, err := q.repo.GetByID(ctx, params.ID)
	if err != nil {
		return GetProviderResult{}, err
	}
	return GetProviderResult{Provider: p}, nil
}
