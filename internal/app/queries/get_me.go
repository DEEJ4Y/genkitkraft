package queries

import (
	"context"

	"github.com/DEEJ4Y/genkitkraft/internal/ports/session"
)

type GetMeParams struct {
	Token string
}

type GetMeResult struct {
	Username string
}

type GetMeQuery struct {
	sessionStore session.Store
}

func NewGetMeQuery(sessionStore session.Store) *GetMeQuery {
	return &GetMeQuery{sessionStore: sessionStore}
}

func (q *GetMeQuery) Execute(ctx context.Context, params GetMeParams) (GetMeResult, error) {
	username, err := q.sessionStore.Validate(ctx, params.Token)
	if err != nil {
		return GetMeResult{}, err
	}
	return GetMeResult{Username: username}, nil
}
