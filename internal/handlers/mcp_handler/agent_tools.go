package mcphandler

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/DEEJ4Y/genkitkraft/internal/app/commands"
	"github.com/DEEJ4Y/genkitkraft/internal/app/queries"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/agent"
)

// --- Input/Output types ---

type ListAgentsInput struct {
	Limit  int `json:"limit" jsonschema:"max number of agents to return (default 20, max 100)"`
	Offset int `json:"offset" jsonschema:"offset for pagination"`
}

type AgentOutput struct {
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	ProviderID         string    `json:"provider_id"`
	ModelID            string    `json:"model_id"`
	SystemPromptID     string    `json:"system_prompt_id,omitempty"`
	TemperatureEnabled bool      `json:"temperature_enabled"`
	Temperature        float64   `json:"temperature"`
	TopPEnabled        bool      `json:"top_p_enabled"`
	TopP               float64   `json:"top_p"`
	TopKEnabled        bool      `json:"top_k_enabled"`
	TopK               int       `json:"top_k"`
	ProviderName       string    `json:"provider_name,omitempty"`
	ProviderType       string    `json:"provider_type,omitempty"`
	SystemPromptName   string    `json:"system_prompt_name,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type ListAgentsOutput struct {
	Agents []AgentOutput `json:"agents"`
	Total  int           `json:"total"`
}

type GetAgentInput struct {
	ID string `json:"id" jsonschema:"agent ID"`
}

type CreateAgentInput struct {
	Name               string   `json:"name" jsonschema:"agent name (required)"`
	ProviderID         string   `json:"provider_id" jsonschema:"LLM provider ID (required)"`
	ModelID            string   `json:"model_id" jsonschema:"model identifier (required)"`
	SystemPromptID     string   `json:"system_prompt_id,omitempty" jsonschema:"system prompt ID"`
	TemperatureEnabled *bool    `json:"temperature_enabled,omitempty" jsonschema:"enable temperature sampling"`
	Temperature        *float64 `json:"temperature,omitempty" jsonschema:"temperature value (0-2)"`
	TopPEnabled        *bool    `json:"top_p_enabled,omitempty" jsonschema:"enable top-p sampling"`
	TopP               *float64 `json:"top_p,omitempty" jsonschema:"top-p value (0-1)"`
	TopKEnabled        *bool    `json:"top_k_enabled,omitempty" jsonschema:"enable top-k sampling"`
	TopK               *int     `json:"top_k,omitempty" jsonschema:"top-k value"`
}

type UpdateAgentInput struct {
	ID                 string   `json:"id" jsonschema:"agent ID (required)"`
	Name               *string  `json:"name,omitempty" jsonschema:"new agent name"`
	ProviderID         *string  `json:"provider_id,omitempty" jsonschema:"new provider ID"`
	ModelID            *string  `json:"model_id,omitempty" jsonschema:"new model ID"`
	SystemPromptID     *string  `json:"system_prompt_id,omitempty" jsonschema:"new system prompt ID (empty string to clear)"`
	TemperatureEnabled *bool    `json:"temperature_enabled,omitempty" jsonschema:"enable temperature sampling"`
	Temperature        *float64 `json:"temperature,omitempty" jsonschema:"temperature value"`
	TopPEnabled        *bool    `json:"top_p_enabled,omitempty" jsonschema:"enable top-p sampling"`
	TopP               *float64 `json:"top_p,omitempty" jsonschema:"top-p value"`
	TopKEnabled        *bool    `json:"top_k_enabled,omitempty" jsonschema:"enable top-k sampling"`
	TopK               *int     `json:"top_k,omitempty" jsonschema:"top-k value"`
}

type DeleteAgentInput struct {
	ID string `json:"id" jsonschema:"agent ID to delete"`
}

// --- Tool registration ---

func (h *Handler) registerAgentTools(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "agents_list",
		Description: "List all agents with pagination.",
	}, h.listAgents)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "agents_get",
		Description: "Get a single agent by ID.",
	}, h.getAgent)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "agents_create",
		Description: "Create a new agent with a provider, model, and optional system prompt.",
	}, h.createAgent)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "agents_update",
		Description: "Update an existing agent. Only provided fields are changed.",
	}, h.updateAgent)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "agents_delete",
		Description: "Delete an agent by ID.",
	}, h.deleteAgent)
}

// --- Tool handlers ---

func (h *Handler) listAgents(ctx context.Context, _ *mcp.CallToolRequest, input ListAgentsInput) (*mcp.CallToolResult, ListAgentsOutput, error) {
	result, err := h.agentApp.Queries.ListAgents.Execute(ctx, queries.ListAgentsParams{
		Limit:  input.Limit,
		Offset: input.Offset,
	})
	if err != nil {
		return nil, ListAgentsOutput{}, fmt.Errorf("list agents failed: %w", err)
	}
	agents := make([]AgentOutput, len(result.Agents))
	for i, a := range result.Agents {
		agents[i] = toAgentOutput(a)
	}
	return nil, ListAgentsOutput{Agents: agents, Total: result.Total}, nil
}

func (h *Handler) getAgent(ctx context.Context, _ *mcp.CallToolRequest, input GetAgentInput) (*mcp.CallToolResult, AgentOutput, error) {
	result, err := h.agentApp.Queries.GetAgent.Execute(ctx, queries.GetAgentParams{ID: input.ID})
	if err != nil {
		return nil, AgentOutput{}, fmt.Errorf("get agent failed: %w", err)
	}
	return nil, toAgentOutput(result.Agent), nil
}

func (h *Handler) createAgent(ctx context.Context, _ *mcp.CallToolRequest, input CreateAgentInput) (*mcp.CallToolResult, AgentOutput, error) {
	result, err := h.agentApp.Commands.CreateAgent.Execute(ctx, commands.CreateAgentParams{
		Name:               input.Name,
		ProviderID:         input.ProviderID,
		ModelID:            input.ModelID,
		SystemPromptID:     input.SystemPromptID,
		TemperatureEnabled: input.TemperatureEnabled,
		Temperature:        input.Temperature,
		TopPEnabled:        input.TopPEnabled,
		TopP:               input.TopP,
		TopKEnabled:        input.TopKEnabled,
		TopK:               input.TopK,
	})
	if err != nil {
		return nil, AgentOutput{}, fmt.Errorf("create agent failed: %w", err)
	}
	return nil, toAgentOutput(result.Agent), nil
}

func (h *Handler) updateAgent(ctx context.Context, _ *mcp.CallToolRequest, input UpdateAgentInput) (*mcp.CallToolResult, AgentOutput, error) {
	result, err := h.agentApp.Commands.UpdateAgent.Execute(ctx, commands.UpdateAgentParams{
		ID:                 input.ID,
		Name:               input.Name,
		ProviderID:         input.ProviderID,
		ModelID:            input.ModelID,
		SystemPromptID:     input.SystemPromptID,
		TemperatureEnabled: input.TemperatureEnabled,
		Temperature:        input.Temperature,
		TopPEnabled:        input.TopPEnabled,
		TopP:               input.TopP,
		TopKEnabled:        input.TopKEnabled,
		TopK:               input.TopK,
	})
	if err != nil {
		return nil, AgentOutput{}, fmt.Errorf("update agent failed: %w", err)
	}
	return nil, toAgentOutput(result.Agent), nil
}

func (h *Handler) deleteAgent(ctx context.Context, _ *mcp.CallToolRequest, input DeleteAgentInput) (*mcp.CallToolResult, any, error) {
	err := h.agentApp.Commands.DeleteAgent.Execute(ctx, commands.DeleteAgentParams{ID: input.ID})
	if err != nil {
		return nil, nil, fmt.Errorf("delete agent failed: %w", err)
	}
	return nil, map[string]string{"status": "deleted"}, nil
}

// --- Helpers ---

func toAgentOutput(a *agent.Agent) AgentOutput {
	return AgentOutput{
		ID:                 a.ID,
		Name:               a.Name,
		ProviderID:         a.ProviderID,
		ModelID:            a.ModelID,
		SystemPromptID:     a.SystemPromptID,
		TemperatureEnabled: a.TemperatureEnabled,
		Temperature:        a.Temperature,
		TopPEnabled:        a.TopPEnabled,
		TopP:               a.TopP,
		TopKEnabled:        a.TopKEnabled,
		TopK:               a.TopK,
		ProviderName:       a.ProviderName,
		ProviderType:       a.ProviderType,
		SystemPromptName:   a.SystemPromptName,
		CreatedAt:          a.CreatedAt,
		UpdatedAt:          a.UpdatedAt,
	}
}
