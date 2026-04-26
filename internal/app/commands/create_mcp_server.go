package commands

import (
	"context"

	"github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	mcpserver "github.com/DEEJ4Y/genkitkraft/internal/domain/mcp_server"
	mcpserverrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/mcp_server_repo"
)

type CreateMcpServerParams struct {
	Name      string
	Transport string
	URL       string
	Headers   []mcpserver.McpServerHeader
}

type CreateMcpServerResult struct {
	McpServer *mcpserver.McpServer
}

type CreateMcpServerCommand struct {
	repo mcpserverrepo.McpServerRepository
}

func NewCreateMcpServerCommand(repo mcpserverrepo.McpServerRepository) *CreateMcpServerCommand {
	return &CreateMcpServerCommand{repo: repo}
}

func (c *CreateMcpServerCommand) Execute(ctx context.Context, params CreateMcpServerParams) (CreateMcpServerResult, error) {
	if params.Name == "" {
		return CreateMcpServerResult{}, errors.NewAppError(errors.InvalidInput, "name is required")
	}
	if params.URL == "" {
		return CreateMcpServerResult{}, errors.NewAppError(errors.InvalidInput, "url is required")
	}

	transport := params.Transport
	if transport == "" {
		transport = "sse"
	}

	headers := params.Headers
	if headers == nil {
		headers = []mcpserver.McpServerHeader{}
	}

	s := &mcpserver.McpServer{
		Name:      params.Name,
		Transport: transport,
		URL:       params.URL,
		Headers:   headers,
	}

	if err := c.repo.Create(ctx, s); err != nil {
		return CreateMcpServerResult{}, err
	}

	return CreateMcpServerResult{McpServer: s}, nil
}
