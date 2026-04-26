package mcpserver

import "time"

// McpServerHeader represents a key-value pair for an HTTP header.
type McpServerHeader struct {
	Name  string
	Value string
}

// McpServer represents a remote MCP server configuration.
type McpServer struct {
	ID        string
	Name      string
	Transport string // "sse" or "streamableHttp"
	URL       string
	Headers   []McpServerHeader
	CreatedAt time.Time
	UpdatedAt time.Time
}
