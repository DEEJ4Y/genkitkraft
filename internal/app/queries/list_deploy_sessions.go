package queries

import (
	"context"

	"github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/playground"
	agentrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/agent_repo"
	playgroundrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/playground_repo"
)

type ListDeploySessionsParams struct {
	AgentID string
	Limit   int
	Offset  int
}

type ListDeploySessionsResult struct {
	Sessions []*playground.Session
	Total    int
	// Limit and Offset are the effective values after defaults and clamping.
	Limit  int
	Offset int
}

// ListDeploySessionsQuery lists one page of an agent's sessions for the deploy API. Unlike the
// playground list it verifies the agent exists (unknown agent is NotFound, not an empty list) and
// paginates.
type ListDeploySessionsQuery struct {
	repo      playgroundrepo.PlaygroundRepository
	agentRepo agentrepo.AgentRepository
}

func NewListDeploySessionsQuery(repo playgroundrepo.PlaygroundRepository, agentRepo agentrepo.AgentRepository) *ListDeploySessionsQuery {
	return &ListDeploySessionsQuery{repo: repo, agentRepo: agentRepo}
}

func (q *ListDeploySessionsQuery) Execute(ctx context.Context, params ListDeploySessionsParams) (ListDeploySessionsResult, error) {
	if params.AgentID == "" {
		return ListDeploySessionsResult{}, errors.NewAppError(errors.InvalidInput, "agent ID is required")
	}

	if _, err := q.agentRepo.GetByID(ctx, params.AgentID); err != nil {
		return ListDeploySessionsResult{}, err
	}

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

	total, err := q.repo.CountSessionsByAgent(ctx, params.AgentID)
	if err != nil {
		return ListDeploySessionsResult{}, err
	}

	sessions, err := q.repo.ListSessionsByAgentPaged(ctx, params.AgentID, limit, offset)
	if err != nil {
		return ListDeploySessionsResult{}, err
	}

	return ListDeploySessionsResult{
		Sessions: sessions,
		Total:    total,
		Limit:    limit,
		Offset:   offset,
	}, nil
}
