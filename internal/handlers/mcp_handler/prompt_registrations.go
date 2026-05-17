package mcphandler

import (
	"context"
	_ "embed"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed prompts/create_agent.md
var createAgentPromptContent string

//go:embed prompts/backup.md
var backupPromptContent string

//go:embed prompts/restore.md
var restorePromptContent string

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

	s.AddPrompt(&mcp.Prompt{
		Name:        "backup",
		Description: "Guide for backing up all GenKitKraft configurations (tools, MCP servers, prompts, agents) to a structured markdown file.",
	}, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		return &mcp.GetPromptResult{
			Description: "Step-by-step guide for exporting all GenKitKraft configurations to a backup markdown file.",
			Messages: []*mcp.PromptMessage{
				{
					Role:    "user",
					Content: &mcp.TextContent{Text: backupPromptContent},
				},
			},
		}, nil
	})

	s.AddPrompt(&mcp.Prompt{
		Name:        "restore",
		Description: "Guide for restoring GenKitKraft configurations from a backup markdown file, with conflict detection and resolution.",
	}, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		return &mcp.GetPromptResult{
			Description: "Step-by-step guide for restoring GenKitKraft configurations from a backup file with conflict resolution.",
			Messages: []*mcp.PromptMessage{
				{
					Role:    "user",
					Content: &mcp.TextContent{Text: restorePromptContent},
				},
			},
		}, nil
	})
}
