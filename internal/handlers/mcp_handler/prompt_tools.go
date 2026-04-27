package mcphandler

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/DEEJ4Y/genkitkraft/internal/app/commands"
	"github.com/DEEJ4Y/genkitkraft/internal/app/queries"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/prompt"
)

// --- Input/Output types ---

type ListPromptsInput struct {
	Limit  int `json:"limit" jsonschema:"max number of prompts to return (default 20, max 100)"`
	Offset int `json:"offset" jsonschema:"offset for pagination"`
}

type PromptOutput struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ListPromptsOutput struct {
	Prompts []PromptOutput `json:"prompts"`
	Total   int            `json:"total"`
}

type GetPromptInput struct {
	ID string `json:"id" jsonschema:"prompt ID"`
}

type CreatePromptInput struct {
	Name    string `json:"name" jsonschema:"prompt name (required)"`
	Content string `json:"content" jsonschema:"prompt content / system instructions (required)"`
}

type UpdatePromptInput struct {
	ID      string  `json:"id" jsonschema:"prompt ID (required)"`
	Name    *string `json:"name,omitempty" jsonschema:"new prompt name"`
	Content *string `json:"content,omitempty" jsonschema:"new prompt content"`
}

type DeletePromptInput struct {
	ID string `json:"id" jsonschema:"prompt ID to delete"`
}

// --- Tool registration ---

func (h *Handler) registerPromptTools(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "prompts_list",
		Description: "List all system prompts with pagination.",
	}, h.listPrompts)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "prompts_get",
		Description: "Get a system prompt by ID.",
	}, h.getPrompt)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "prompts_create",
		Description: "Create a new system prompt.",
	}, h.createPrompt)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "prompts_update",
		Description: "Update an existing system prompt. Only provided fields are changed.",
	}, h.updatePrompt)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "prompts_delete",
		Description: "Delete a system prompt by ID.",
	}, h.deletePrompt)
}

// --- Tool handlers ---

func (h *Handler) listPrompts(ctx context.Context, _ *mcp.CallToolRequest, input ListPromptsInput) (*mcp.CallToolResult, ListPromptsOutput, error) {
	result, err := h.promptApp.Queries.ListPrompts.Execute(ctx, queries.ListPromptsParams{
		Limit:  input.Limit,
		Offset: input.Offset,
	})
	if err != nil {
		return nil, ListPromptsOutput{}, fmt.Errorf("list prompts failed: %w", err)
	}
	prompts := make([]PromptOutput, len(result.Prompts))
	for i, p := range result.Prompts {
		prompts[i] = toPromptOutput(p)
	}
	return nil, ListPromptsOutput{Prompts: prompts, Total: result.Total}, nil
}

func (h *Handler) getPrompt(ctx context.Context, _ *mcp.CallToolRequest, input GetPromptInput) (*mcp.CallToolResult, PromptOutput, error) {
	result, err := h.promptApp.Queries.GetPrompt.Execute(ctx, queries.GetPromptParams{ID: input.ID})
	if err != nil {
		return nil, PromptOutput{}, fmt.Errorf("get prompt failed: %w", err)
	}
	return nil, toPromptOutput(result.Prompt), nil
}

func (h *Handler) createPrompt(ctx context.Context, _ *mcp.CallToolRequest, input CreatePromptInput) (*mcp.CallToolResult, PromptOutput, error) {
	result, err := h.promptApp.Commands.CreatePrompt.Execute(ctx, commands.CreatePromptParams{
		Name:    input.Name,
		Content: input.Content,
	})
	if err != nil {
		return nil, PromptOutput{}, fmt.Errorf("create prompt failed: %w", err)
	}
	return nil, toPromptOutput(result.Prompt), nil
}

func (h *Handler) updatePrompt(ctx context.Context, _ *mcp.CallToolRequest, input UpdatePromptInput) (*mcp.CallToolResult, PromptOutput, error) {
	result, err := h.promptApp.Commands.UpdatePrompt.Execute(ctx, commands.UpdatePromptParams{
		ID:      input.ID,
		Name:    input.Name,
		Content: input.Content,
	})
	if err != nil {
		return nil, PromptOutput{}, fmt.Errorf("update prompt failed: %w", err)
	}
	return nil, toPromptOutput(result.Prompt), nil
}

func (h *Handler) deletePrompt(ctx context.Context, _ *mcp.CallToolRequest, input DeletePromptInput) (*mcp.CallToolResult, any, error) {
	err := h.promptApp.Commands.DeletePrompt.Execute(ctx, commands.DeletePromptParams{ID: input.ID})
	if err != nil {
		return nil, nil, fmt.Errorf("delete prompt failed: %w", err)
	}
	return nil, map[string]string{"status": "deleted"}, nil
}

func toPromptOutput(p *prompt.Prompt) PromptOutput {
	return PromptOutput{
		ID:        p.ID,
		Name:      p.Name,
		Content:   p.Content,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}
