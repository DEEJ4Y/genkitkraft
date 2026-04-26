package queries

import (
	"context"

	mcpserver "github.com/DEEJ4Y/genkitkraft/internal/domain/mcp_server"
	mcpserverrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/mcp_server_repo"
)

type GetMcpServerParams struct {
	ID string
}

type GetMcpServerResult struct {
	McpServer *mcpserver.McpServer
}

type GetMcpServerQuery struct {
	repo mcpserverrepo.McpServerRepository
}

func NewGetMcpServerQuery(repo mcpserverrepo.McpServerRepository) *GetMcpServerQuery {
	return &GetMcpServerQuery{repo: repo}
}

func (q *GetMcpServerQuery) Execute(ctx context.Context, params GetMcpServerParams) (GetMcpServerResult, error) {
	s, err := q.repo.GetByID(ctx, params.ID)
	if err != nil {
		return GetMcpServerResult{}, err
	}
	return GetMcpServerResult{McpServer: s}, nil
}
