package mcphandler

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// --- Output types ---

type HealthOutput struct {
	Status string `json:"status"`
}

// --- Tool registration ---

func (h *Handler) registerHealthTools(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "health_liveness",
		Description: "Check if the server is alive.",
	}, h.healthLiveness)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "health_readiness",
		Description: "Check if the server is ready to accept requests.",
	}, h.healthReadiness)
}

// --- Tool handlers ---

func (h *Handler) healthLiveness(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, HealthOutput, error) {
	return nil, HealthOutput{Status: "ok"}, nil
}

func (h *Handler) healthReadiness(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, HealthOutput, error) {
	return nil, HealthOutput{Status: "ok"}, nil
}
