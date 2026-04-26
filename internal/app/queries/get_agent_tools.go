package queries

import (
	"context"

	agenttoolrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/agent_tool_repo"
)

type GetAgentToolsParams struct {
	AgentID string
}

type GetAgentToolsResult struct {
	Config agenttoolrepo.AgentToolConfig
}

type GetAgentToolsQuery struct {
	repo agenttoolrepo.AgentToolRepository
}

func NewGetAgentToolsQuery(repo agenttoolrepo.AgentToolRepository) *GetAgentToolsQuery {
	return &GetAgentToolsQuery{repo: repo}
}

func (q *GetAgentToolsQuery) Execute(ctx context.Context, params GetAgentToolsParams) (GetAgentToolsResult, error) {
	config, err := q.repo.GetByAgentID(ctx, params.AgentID)
	if err != nil {
		return GetAgentToolsResult{}, err
	}
	return GetAgentToolsResult{Config: config}, nil
}
