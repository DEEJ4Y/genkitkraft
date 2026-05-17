package commands

import (
	"context"

	agenttoolrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/agent_tool_repo"
)

type UpdateAgentToolsParams struct {
	AgentID        string
	HttpToolIDs    []string
	McpServers     []agenttoolrepo.McpServerToolConfig
	BuiltInToolIDs []string
}

type UpdateAgentToolsResult struct {
	Config agenttoolrepo.AgentToolConfig
}

type UpdateAgentToolsCommand struct {
	repo agenttoolrepo.AgentToolRepository
}

func NewUpdateAgentToolsCommand(repo agenttoolrepo.AgentToolRepository) *UpdateAgentToolsCommand {
	return &UpdateAgentToolsCommand{repo: repo}
}

func (c *UpdateAgentToolsCommand) Execute(ctx context.Context, params UpdateAgentToolsParams) (UpdateAgentToolsResult, error) {
	config := agenttoolrepo.AgentToolConfig{
		AgentID:        params.AgentID,
		HttpToolIDs:    params.HttpToolIDs,
		McpServers:     params.McpServers,
		BuiltInToolIDs: params.BuiltInToolIDs,
	}

	if config.HttpToolIDs == nil {
		config.HttpToolIDs = []string{}
	}
	if config.McpServers == nil {
		config.McpServers = []agenttoolrepo.McpServerToolConfig{}
	}
	if config.BuiltInToolIDs == nil {
		config.BuiltInToolIDs = []string{}
	}

	if err := c.repo.Save(ctx, config); err != nil {
		return UpdateAgentToolsResult{}, err
	}

	saved, err := c.repo.GetByAgentID(ctx, params.AgentID)
	if err != nil {
		return UpdateAgentToolsResult{}, err
	}

	return UpdateAgentToolsResult{Config: saved}, nil
}
