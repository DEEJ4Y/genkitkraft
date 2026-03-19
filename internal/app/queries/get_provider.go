package queries

import (
	"context"

	"github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/provider"
	"github.com/DEEJ4Y/genkitkraft/internal/ports/encryptor"
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
	enc  encryptor.Encryptor
}

func NewGetProviderQuery(repo providerrepo.ProviderRepository, enc encryptor.Encryptor) *GetProviderQuery {
	return &GetProviderQuery{repo: repo, enc: enc}
}

func (q *GetProviderQuery) Execute(ctx context.Context, params GetProviderParams) (GetProviderResult, error) {
	p, err := q.repo.GetByID(ctx, params.ID)
	if err != nil {
		return GetProviderResult{}, err
	}

	p.APIKey, err = q.enc.Decrypt(p.APIKey)
	if err != nil {
		return GetProviderResult{}, errors.NewAppErrorf(errors.Internal, "decrypting api key: %v", err)
	}

	return GetProviderResult{Provider: p}, nil
}
