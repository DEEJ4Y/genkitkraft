package mcpdiscovery

import "context"

// McpTool represents a tool exposed by an MCP server.
type McpTool struct {
	Name        string
	Description string
}

// McpDiscovery defines the interface for discovering tools from MCP servers.
type McpDiscovery interface {
	// ListTools connects to an MCP server and returns its available tools.
	// It tries the given transport first and falls back to the alternate
	// transport (SSE ↔ Streamable HTTP) if the initial connection fails.
	ListTools(ctx context.Context, serverName, transport, url string, headers map[string]string) ([]McpTool, error)
}
