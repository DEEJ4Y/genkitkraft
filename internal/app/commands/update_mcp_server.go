package commands

import (
	"context"

	mcpserver "github.com/DEEJ4Y/genkitkraft/internal/domain/mcp_server"
	mcpserverrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/mcp_server_repo"
)

type UpdateMcpServerParams struct {
	ID        string
	Name      *string
	Transport *string
	URL       *string
	Headers   *[]mcpserver.McpServerHeader
}

type UpdateMcpServerResult struct {
	McpServer *mcpserver.McpServer
}

type UpdateMcpServerCommand struct {
	repo mcpserverrepo.McpServerRepository
}

func NewUpdateMcpServerCommand(repo mcpserverrepo.McpServerRepository) *UpdateMcpServerCommand {
	return &UpdateMcpServerCommand{repo: repo}
}

func (c *UpdateMcpServerCommand) Execute(ctx context.Context, params UpdateMcpServerParams) (UpdateMcpServerResult, error) {
	s, err := c.repo.GetByID(ctx, params.ID)
	if err != nil {
		return UpdateMcpServerResult{}, err
	}

	if params.Name != nil {
		s.Name = *params.Name
	}
	if params.Transport != nil {
		s.Transport = *params.Transport
	}
	if params.URL != nil {
		s.URL = *params.URL
	}
	if params.Headers != nil {
		s.Headers = *params.Headers
	}

	if err := c.repo.Update(ctx, s); err != nil {
		return UpdateMcpServerResult{}, err
	}

	return UpdateMcpServerResult{McpServer: s}, nil
}
