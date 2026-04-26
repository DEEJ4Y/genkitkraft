package commands

import (
	"context"

	mcpserverrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/mcp_server_repo"
)

type DeleteMcpServerParams struct {
	ID string
}

type DeleteMcpServerCommand struct {
	repo mcpserverrepo.McpServerRepository
}

func NewDeleteMcpServerCommand(repo mcpserverrepo.McpServerRepository) *DeleteMcpServerCommand {
	return &DeleteMcpServerCommand{repo: repo}
}

func (c *DeleteMcpServerCommand) Execute(ctx context.Context, params DeleteMcpServerParams) error {
	return c.repo.Delete(ctx, params.ID)
}
