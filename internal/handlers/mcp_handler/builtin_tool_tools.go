package mcphandler

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	builtintool "github.com/DEEJ4Y/genkitkraft/internal/domain/builtin_tool"
)

// --- Input/Output types ---

type ListBuiltInToolsInput struct{}

type BuiltInToolOutput struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ListBuiltInToolsOutput struct {
	BuiltInTools []BuiltInToolOutput `json:"built_in_tools"`
}

// --- Tool registration ---

func (h *Handler) registerBuiltInToolTools(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "built_in_tools_list",
		Description: "List all available built-in tools.",
	}, h.listBuiltInTools)
}

func (h *Handler) listBuiltInTools(ctx context.Context, _ *mcp.CallToolRequest, _ ListBuiltInToolsInput) (*mcp.CallToolResult, ListBuiltInToolsOutput, error) {
	result, err := h.builtInToolApp.Queries.ListBuiltInTools.Execute(ctx)
	if err != nil {
		return nil, ListBuiltInToolsOutput{}, err
	}

	tools := make([]BuiltInToolOutput, len(result.BuiltInTools))
	for i, t := range result.BuiltInTools {
		tools[i] = toBuiltInToolMcpOutput(t)
	}

	return nil, ListBuiltInToolsOutput{BuiltInTools: tools}, nil
}

func toBuiltInToolMcpOutput(t builtintool.BuiltInTool) BuiltInToolOutput {
	return BuiltInToolOutput{
		ID:          t.ID,
		Name:        t.Name,
		Description: t.Description,
	}
}
