package agenttoolrepo

import "context"

// McpServerToolConfig represents the tool selection for a single MCP server.
type McpServerToolConfig struct {
	McpServerID string
	SelectAll   bool
	ToolNames   []string
}

// AgentToolConfig represents an agent's complete tool configuration.
type AgentToolConfig struct {
	AgentID    string
	HttpToolIDs []string
	McpServers []McpServerToolConfig
}

// AgentToolRepository defines the contract for managing agent-tool associations.
type AgentToolRepository interface {
	// GetByAgentID returns the tool configuration for an agent.
	GetByAgentID(ctx context.Context, agentID string) (AgentToolConfig, error)

	// Save replaces the entire tool configuration for an agent.
	Save(ctx context.Context, config AgentToolConfig) error
}
