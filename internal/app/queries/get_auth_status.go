package queries

import "context"

type GetAuthStatusParams struct{}

type GetAuthStatusResult struct {
	Required bool
}

type GetAuthStatusQuery struct {
	authRequired bool
}

func NewGetAuthStatusQuery(authRequired bool) *GetAuthStatusQuery {
	return &GetAuthStatusQuery{authRequired: authRequired}
}

func (q *GetAuthStatusQuery) Execute(_ context.Context, _ GetAuthStatusParams) (GetAuthStatusResult, error) {
	return GetAuthStatusResult{Required: q.authRequired}, nil
}
