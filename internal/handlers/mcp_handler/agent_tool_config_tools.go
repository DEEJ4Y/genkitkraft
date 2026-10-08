package mcphandler

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/DEEJ4Y/genkitkraft/internal/app/commands"
	"github.com/DEEJ4Y/genkitkraft/internal/app/queries"
	agenttoolrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/agent_tool_repo"
)

// --- Input/Output types ---

type GetAgentToolConfigInput struct {
	AgentID string `json:"agent_id" jsonschema:"agent ID (required)"`
}

type McpServerToolConfigOutput struct {
	McpServerID string   `json:"mcp_server_id"`
	SelectAll   bool     `json:"select_all"`
	ToolNames   []string `json:"tool_names,omitempty"`
}

type AgentToolConfigOutput struct {
	AgentID        string                      `json:"agent_id"`
	HttpToolIDs    []string                    `json:"http_tool_ids"`
	McpServers     []McpServerToolConfigOutput `json:"mcp_servers"`
	BuiltInToolIDs []string                    `json:"built_in_tool_ids"`
}

type McpServerToolConfigInput struct {
	McpServerID string   `json:"mcp_server_id" jsonschema:"MCP server ID"`
	SelectAll   bool     `json:"select_all" jsonschema:"whether to include all tools from this server"`
	ToolNames   []string `json:"tool_names,omitempty" jsonschema:"specific tool names to include (if select_all is false)"`
}

type UpdateAgentToolConfigInput struct {
	AgentID        string                     `json:"agent_id" jsonschema:"agent ID (required)"`
	HttpToolIDs    []string                   `json:"http_tool_ids" jsonschema:"list of HTTP tool IDs to assign"`
	McpServers     []McpServerToolConfigInput `json:"mcp_servers" jsonschema:"MCP server tool selections"`
	BuiltInToolIDs []string                   `json:"built_in_tool_ids" jsonschema:"list of built-in tool IDs to enable, e.g. web_fetch (see built_in_tools_list). Omitting this disables all built-in tools"`
}

// --- Tool registration ---

func (h *Handler) registerAgentToolTools(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "agent_tools_get",
		Description: "Get the tool configuration for an agent (which HTTP tools, MCP servers and built-in tools such as web_fetch are assigned).",
	}, h.getAgentToolConfig)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "agent_tools_update",
		Description: "Update the tool configuration for an agent. Replaces the entire configuration (HTTP tools, MCP servers and built-in tools), so include the current values of anything you want to keep; call agent_tools_get first.",
	}, h.updateAgentToolConfig)
}

// --- Tool handlers ---

func (h *Handler) getAgentToolConfig(ctx context.Context, _ *mcp.CallToolRequest, input GetAgentToolConfigInput) (*mcp.CallToolResult, AgentToolConfigOutput, error) {
	result, err := h.agentToolApp.Queries.GetTools.Execute(ctx, queries.GetAgentToolsParams{AgentID: input.AgentID})
	if err != nil {
		return nil, AgentToolConfigOutput{}, fmt.Errorf("get agent tools failed: %w", err)
	}
	return nil, toAgentToolConfigOutput(result.Config), nil
}

func (h *Handler) updateAgentToolConfig(ctx context.Context, _ *mcp.CallToolRequest, input UpdateAgentToolConfigInput) (*mcp.CallToolResult, AgentToolConfigOutput, error) {
	mcpServers := make([]agenttoolrepo.McpServerToolConfig, len(input.McpServers))
	for i, s := range input.McpServers {
		mcpServers[i] = agenttoolrepo.McpServerToolConfig{
			McpServerID: s.McpServerID,
			SelectAll:   s.SelectAll,
			ToolNames:   s.ToolNames,
		}
	}

	result, err := h.agentToolApp.Commands.UpdateTools.Execute(ctx, commands.UpdateAgentToolsParams{
		AgentID:        input.AgentID,
		HttpToolIDs:    input.HttpToolIDs,
		McpServers:     mcpServers,
		BuiltInToolIDs: input.BuiltInToolIDs,
	})
	if err != nil {
		return nil, AgentToolConfigOutput{}, fmt.Errorf("update agent tools failed: %w", err)
	}
	return nil, toAgentToolConfigOutput(result.Config), nil
}

func toAgentToolConfigOutput(cfg agenttoolrepo.AgentToolConfig) AgentToolConfigOutput {
	mcpServers := make([]McpServerToolConfigOutput, len(cfg.McpServers))
	for i, s := range cfg.McpServers {
		mcpServers[i] = McpServerToolConfigOutput{
			McpServerID: s.McpServerID,
			SelectAll:   s.SelectAll,
			ToolNames:   s.ToolNames,
		}
	}
	return AgentToolConfigOutput{
		AgentID:        cfg.AgentID,
		HttpToolIDs:    cfg.HttpToolIDs,
		McpServers:     mcpServers,
		BuiltInToolIDs: cfg.BuiltInToolIDs,
	}
}
