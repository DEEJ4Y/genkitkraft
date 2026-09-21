package mcphandler

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/DEEJ4Y/genkitkraft/internal/app/commands"
	"github.com/DEEJ4Y/genkitkraft/internal/app/queries"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/gap"
)

// --- Input/Output types ---

type ListGapsInput struct {
	AgentID string `json:"agent_id" jsonschema:"agent ID (required)"`
	Status  string `json:"status" jsonschema:"filter by status: open, resolved, or dismissed (optional, omit for no filter)"`
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
	ID      string `json:"id" jsonschema:"gap ID"`
	AgentID string `json:"agent_id" jsonschema:"agent ID (required)"`
}

type ResolveGapInput struct {
	ID      string `json:"id" jsonschema:"gap ID"`
	AgentID string `json:"agent_id" jsonschema:"agent ID (required)"`
}

type DismissGapInput struct {
	ID                string `json:"id" jsonschema:"gap ID"`
	AgentID           string `json:"agent_id" jsonschema:"agent ID (required)"`
	DismissalCategory string `json:"dismissal_category" jsonschema:"reason category: unrelated, insufficient_detail, duplicate, or other (required)"`
	DismissalReason   string `json:"dismissal_reason" jsonschema:"optional free-text dismissal reason"`
}

type ReopenGapInput struct {
	ID      string `json:"id" jsonschema:"gap ID"`
	AgentID string `json:"agent_id" jsonschema:"agent ID (required)"`
}

// --- Tool registration ---
//
// search_gaps/edit_gap are internal to the background dedup pipeline and are
// never exposed as MCP tools — they're not built-in tools an agent (or an
// MCP client) can call. Everything else the REST API supports for gap
// management (list, get, resolve, dismiss, reopen) is exposed here too.

func (h *Handler) registerGapTools(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "gaps_list",
		Description: "List an agent's self-reported gaps (knowledge, capability, or improvement) with pagination.",
	}, h.listGaps)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "gaps_get",
		Description: "Get a single self-reported gap by ID.",
	}, h.getGap)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "gaps_resolve",
		Description: "Mark a gap resolved, clearing any prior dismissal.",
	}, h.resolveGap)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "gaps_dismiss",
		Description: "Dismiss a gap with a reason category. A gap dismissed as \"unrelated\" becomes terminal and can never be reopened.",
	}, h.dismissGap)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "gaps_reopen",
		Description: "Reopen a resolved or dismissed gap. Fails if the gap was dismissed as \"unrelated\", which is terminal.",
	}, h.reopenGap)
}

// --- Tool handlers ---

func (h *Handler) listGaps(ctx context.Context, _ *mcp.CallToolRequest, input ListGapsInput) (*mcp.CallToolResult, ListGapsOutput, error) {
	result, err := h.gapApp.Queries.ListGaps.Execute(ctx, queries.ListGapsParams{
		AgentID: input.AgentID,
		Status:  gap.Status(input.Status),
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
	result, err := h.gapApp.Queries.GetGap.Execute(ctx, queries.GetGapParams{ID: input.ID, AgentID: input.AgentID})
	if err != nil {
		return nil, GapOutput{}, fmt.Errorf("get gap failed: %w", err)
	}
	return nil, toGapOutput(result.GapWithReferences), nil
}

func (h *Handler) resolveGap(ctx context.Context, _ *mcp.CallToolRequest, input ResolveGapInput) (*mcp.CallToolResult, GapOutput, error) {
	if _, err := h.gapApp.Commands.ResolveGap.Execute(ctx, commands.ResolveGapParams{ID: input.ID, AgentID: input.AgentID}); err != nil {
		return nil, GapOutput{}, fmt.Errorf("resolve gap failed: %w", err)
	}
	return h.getGapOutput(ctx, input.ID, input.AgentID)
}

func (h *Handler) dismissGap(ctx context.Context, _ *mcp.CallToolRequest, input DismissGapInput) (*mcp.CallToolResult, GapOutput, error) {
	_, err := h.gapApp.Commands.DismissGap.Execute(ctx, commands.DismissGapParams{
		ID:                input.ID,
		AgentID:           input.AgentID,
		DismissalCategory: input.DismissalCategory,
		DismissalReason:   input.DismissalReason,
	})
	if err != nil {
		return nil, GapOutput{}, fmt.Errorf("dismiss gap failed: %w", err)
	}
	return h.getGapOutput(ctx, input.ID, input.AgentID)
}

func (h *Handler) reopenGap(ctx context.Context, _ *mcp.CallToolRequest, input ReopenGapInput) (*mcp.CallToolResult, GapOutput, error) {
	if _, err := h.gapApp.Commands.ReopenGap.Execute(ctx, commands.ReopenGapParams{ID: input.ID, AgentID: input.AgentID}); err != nil {
		return nil, GapOutput{}, fmt.Errorf("reopen gap failed: %w", err)
	}
	return h.getGapOutput(ctx, input.ID, input.AgentID)
}

// --- Helpers ---

// getGapOutput refetches a gap after a write, mirroring the HTTP handler's
// UpdateGap behavior — write commands only return the bare Gap, but callers
// of this MCP surface get the same references-included shape as gaps_get.
func (h *Handler) getGapOutput(ctx context.Context, id, agentID string) (*mcp.CallToolResult, GapOutput, error) {
	result, err := h.gapApp.Queries.GetGap.Execute(ctx, queries.GetGapParams{ID: id, AgentID: agentID})
	if err != nil {
		return nil, GapOutput{}, fmt.Errorf("get gap failed: %w", err)
	}
	return nil, toGapOutput(result.GapWithReferences), nil
}

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
