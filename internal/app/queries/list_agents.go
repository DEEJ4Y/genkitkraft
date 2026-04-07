package queries

import (
	"context"

	"github.com/DEEJ4Y/genkitkraft/internal/domain/agent"
	agentrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/agent_repo"
)

type ListAgentsParams struct {
	Limit  int
	Offset int
}

type ListAgentsResult struct {
	Agents []*agent.Agent
	Total  int
}

type ListAgentsQuery struct {
	repo agentrepo.AgentRepository
}

func NewListAgentsQuery(repo agentrepo.AgentRepository) *ListAgentsQuery {
	return &ListAgentsQuery{repo: repo}
}

func (q *ListAgentsQuery) Execute(ctx context.Context, params ListAgentsParams) (ListAgentsResult, error) {
	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	offset := params.Offset
	if offset < 0 {
		offset = 0
	}

	total, err := q.repo.Count(ctx)
	if err != nil {
		return ListAgentsResult{}, err
	}

	agents, err := q.repo.List(ctx, limit, offset)
	if err != nil {
		return ListAgentsResult{}, err
	}

	return ListAgentsResult{
		Agents: agents,
		Total:  total,
	}, nil
}
