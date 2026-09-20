package mcphandler

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/DEEJ4Y/genkitkraft/internal/app/queries"
)

// --- Input/Output types ---

type ListGapsInput struct {
	AgentID string `json:"agent_id" jsonschema:"agent ID (required)"`
	Limit   int    `json:"limit" jsonschema:"max number of gaps to return (default 20, max 100)"`
	Offset  int    `json:"offset" jsonschema:"offset for pagination"`
}

type GapReferenceOutput struct {
	SessionID string `json:"session_id,omitempty"`
	MessageID string `json:"message_id,omitempty"`
}

type GapOutput struct {
	ID                  string               `json:"id"`
	AgentID             string               `json:"agent_id"`
	Category            string               `json:"category"`
	Context             string               `json:"context"`
	Details             string               `json:"details"`
	SuggestedResolution string               `json:"suggested_resolution,omitempty"`
	Status              string               `json:"status"`
	DismissalCategory   string               `json:"dismissal_category,omitempty"`
	DismissalReason     string               `json:"dismissal_reason,omitempty"`
	References          []GapReferenceOutput `json:"references"`
	CreatedAt           time.Time            `json:"created_at"`
	UpdatedAt           time.Time            `json:"updated_at"`
}

type ListGapsOutput struct {
	Gaps  []GapOutput `json:"gaps"`
	Total int         `json:"total"`
}

type GetGapInput struct {
	ID string `json:"id" jsonschema:"gap ID"`
}

// --- Tool registration ---
//
// Only read-only tools are registered here. search_gaps/edit_gap are
// internal to the background dedup pipeline and are never exposed as MCP
// tools — they're not built-in tools an agent (or an MCP client) can call.

func (h *Handler) registerGapTools(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "gaps_list",
		Description: "List an agent's self-reported gaps (knowledge, capability, or improvement) with pagination.",
	}, h.listGaps)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "gaps_get",
		Description: "Get a single self-reported gap by ID.",
	}, h.getGap)
}

// --- Tool handlers ---

func (h *Handler) listGaps(ctx context.Context, _ *mcp.CallToolRequest, input ListGapsInput) (*mcp.CallToolResult, ListGapsOutput, error) {
	result, err := h.gapApp.Queries.ListGaps.Execute(ctx, queries.ListGapsParams{
		AgentID: input.AgentID,
		Limit:   input.Limit,
		Offset:  input.Offset,
	})
	if err != nil {
		return nil, ListGapsOutput{}, fmt.Errorf("list gaps failed: %w", err)
	}
	gaps := make([]GapOutput, len(result.Gaps))
	for i, g := range result.Gaps {
		gaps[i] = toGapOutput(g)
	}
	return nil, ListGapsOutput{Gaps: gaps, Total: result.Total}, nil
}

func (h *Handler) getGap(ctx context.Context, _ *mcp.CallToolRequest, input GetGapInput) (*mcp.CallToolResult, GapOutput, error) {
	result, err := h.gapApp.Queries.GetGap.Execute(ctx, queries.GetGapParams{ID: input.ID})
	if err != nil {
		return nil, GapOutput{}, fmt.Errorf("get gap failed: %w", err)
	}
	return nil, toGapOutput(result.GapWithReferences), nil
}

// --- Helpers ---

func toGapOutput(g queries.GapWithReferences) GapOutput {
	refs := make([]GapReferenceOutput, len(g.References))
	for i, ref := range g.References {
		refs[i] = GapReferenceOutput{SessionID: ref.SessionID, MessageID: ref.MessageID}
	}
	out := GapOutput{
		ID:                  g.Gap.ID,
		AgentID:             g.Gap.AgentID,
		Category:            string(g.Gap.Category),
		Context:             g.Gap.Context,
		Details:             g.Gap.Details,
		SuggestedResolution: g.Gap.SuggestedResolution,
		Status:              string(g.Gap.Status),
		DismissalCategory:   g.Gap.DismissalCategory,
		DismissalReason:     g.Gap.DismissalReason,
		References:          refs,
		CreatedAt:           g.Gap.CreatedAt,
		UpdatedAt:           g.Gap.UpdatedAt,
	}
	return out
}
