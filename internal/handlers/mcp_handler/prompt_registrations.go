package mcphandler

import (
	"context"
	_ "embed"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed prompts/create_agent.md
var createAgentPromptContent string

func (h *Handler) registerPrompts(s *mcp.Server) {
	s.AddPrompt(&mcp.Prompt{
		Name:        "create-agent",
		Description: "Comprehensive guide for creating and configuring GenKitKraft agents via MCP tools. Includes tool reference, step-by-step workflow, and examples.",
	}, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		return &mcp.GetPromptResult{
			Description: "Step-by-step guide for creating GenKitKraft agents with all available MCP tools.",
			Messages: []*mcp.PromptMessage{
				{
					Role:    "user",
					Content: &mcp.TextContent{Text: createAgentPromptContent},
				},
			},
		}, nil
	})
}
