package queries

import (
	"context"

	"github.com/DEEJ4Y/genkitkraft/internal/domain/provider"
)

type ListProviderTypesParams struct{}

type ListProviderTypesResult struct {
	ProviderTypes []provider.ProviderTypeInfo
}

type ListProviderTypesQuery struct{}

func NewListProviderTypesQuery() *ListProviderTypesQuery {
	return &ListProviderTypesQuery{}
}

func (q *ListProviderTypesQuery) Execute(_ context.Context, _ ListProviderTypesParams) (ListProviderTypesResult, error) {
	return ListProviderTypesResult{ProviderTypes: provider.ProviderTypeMeta}, nil
}
