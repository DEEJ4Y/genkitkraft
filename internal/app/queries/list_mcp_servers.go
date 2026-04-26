package queries

import (
	"context"

	mcpserver "github.com/DEEJ4Y/genkitkraft/internal/domain/mcp_server"
	mcpserverrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/mcp_server_repo"
)

type ListMcpServersParams struct {
	Limit  int
	Offset int
}

type ListMcpServersResult struct {
	McpServers []*mcpserver.McpServer
	Total      int
}

type ListMcpServersQuery struct {
	repo mcpserverrepo.McpServerRepository
}

func NewListMcpServersQuery(repo mcpserverrepo.McpServerRepository) *ListMcpServersQuery {
	return &ListMcpServersQuery{repo: repo}
}

func (q *ListMcpServersQuery) Execute(ctx context.Context, params ListMcpServersParams) (ListMcpServersResult, error) {
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
		return ListMcpServersResult{}, err
	}

	servers, err := q.repo.List(ctx, limit, offset)
	if err != nil {
		return ListMcpServersResult{}, err
	}

	return ListMcpServersResult{
		McpServers: servers,
		Total:      total,
	}, nil
}
