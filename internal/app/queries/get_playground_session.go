package queries

import (
	"context"

	apperrors "github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/playground"
	playgroundrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/playground_repo"
)

type GetPlaygroundSessionParams struct {
	SessionID string
	AgentID   string
}

type GetPlaygroundSessionResult struct {
	Session *playground.Session
}

type GetPlaygroundSessionQuery struct {
	repo playgroundrepo.PlaygroundRepository
}

func NewGetPlaygroundSessionQuery(repo playgroundrepo.PlaygroundRepository) *GetPlaygroundSessionQuery {
	return &GetPlaygroundSessionQuery{repo: repo}
}

func (q *GetPlaygroundSessionQuery) Execute(ctx context.Context, params GetPlaygroundSessionParams) (GetPlaygroundSessionResult, error) {
	session, err := q.repo.GetSession(ctx, params.SessionID)
	if err != nil {
		return GetPlaygroundSessionResult{}, err
	}
	if session.AgentID != params.AgentID {
		return GetPlaygroundSessionResult{}, apperrors.NewAppError(apperrors.NotFound, "session not found")
	}
	return GetPlaygroundSessionResult{Session: session}, nil
}
