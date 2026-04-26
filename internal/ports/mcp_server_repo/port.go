package mcpserverrepo

import (
	"context"

	mcpserver "github.com/DEEJ4Y/genkitkraft/internal/domain/mcp_server"
)

// McpServerRepository defines the contract for MCP server configuration persistence.
type McpServerRepository interface {
	List(ctx context.Context, limit, offset int) ([]*mcpserver.McpServer, error)
	Count(ctx context.Context) (int, error)
	GetByID(ctx context.Context, id string) (*mcpserver.McpServer, error)
	Create(ctx context.Context, s *mcpserver.McpServer) error
	Update(ctx context.Context, s *mcpserver.McpServer) error
	Delete(ctx context.Context, id string) error
}
