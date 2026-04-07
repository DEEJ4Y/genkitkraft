package queries

import (
	"context"

	"github.com/DEEJ4Y/genkitkraft/internal/domain/playground"
	playgroundrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/playground_repo"
)

type ListPlaygroundSessionsParams struct {
	AgentID string
}

type ListPlaygroundSessionsResult struct {
	Sessions []*playground.Session
}

type ListPlaygroundSessionsQuery struct {
	repo playgroundrepo.PlaygroundRepository
}

func NewListPlaygroundSessionsQuery(repo playgroundrepo.PlaygroundRepository) *ListPlaygroundSessionsQuery {
	return &ListPlaygroundSessionsQuery{repo: repo}
}

func (q *ListPlaygroundSessionsQuery) Execute(ctx context.Context, params ListPlaygroundSessionsParams) (ListPlaygroundSessionsResult, error) {
	sessions, err := q.repo.ListSessionsByAgent(ctx, params.AgentID)
	if err != nil {
		return ListPlaygroundSessionsResult{}, err
	}
	return ListPlaygroundSessionsResult{Sessions: sessions}, nil
}
