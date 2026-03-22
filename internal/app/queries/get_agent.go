package queries

import (
	"context"

	"github.com/DEEJ4Y/genkitkraft/internal/domain/agent"
	agentrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/agent_repo"
)

type GetAgentParams struct {
	ID string
}

type GetAgentResult struct {
	Agent *agent.Agent
}

type GetAgentQuery struct {
	repo agentrepo.AgentRepository
}

func NewGetAgentQuery(repo agentrepo.AgentRepository) *GetAgentQuery {
	return &GetAgentQuery{repo: repo}
}

func (q *GetAgentQuery) Execute(ctx context.Context, params GetAgentParams) (GetAgentResult, error) {
	a, err := q.repo.GetByID(ctx, params.ID)
	if err != nil {
		return GetAgentResult{}, err
	}
	return GetAgentResult{Agent: a}, nil
}
