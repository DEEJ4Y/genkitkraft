package mcphandler

import (
	"context"
	"testing"

	"github.com/DEEJ4Y/genkitkraft/internal/app"
	"github.com/DEEJ4Y/genkitkraft/internal/app/commands"
	"github.com/DEEJ4Y/genkitkraft/internal/app/queries"
	apperrors "github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/agent"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/gap"
	"github.com/DEEJ4Y/genkitkraft/resources/test/mock"
)

func newTestHandler(repo *mock.GapRepository) *Handler {
	agentRepo := &mock.AgentRepository{GetByIDResult: &agent.Agent{ID: "agent-1"}}
	return &Handler{
		gapApp: &app.GapApp{
			Commands: app.GapCommands{
				ResolveGap: commands.NewResolveGapCommand(repo),
				DismissGap: commands.NewDismissGapCommand(repo),
				ReopenGap:  commands.NewReopenGapCommand(repo),
			},
			Queries: app.GapQueries{
				ListGaps: queries.NewListGapsQuery(repo, agentRepo),
				GetGap:   queries.NewGetGapQuery(repo),
			},
		},
	}
}

func TestListGapsTool_HappyPath(t *testing.T) {
	repo := &mock.GapRepository{
		ListGaps:    []*gap.Gap{{ID: "gap-1", AgentID: "agent-1", Category: gap.CategoryCapability, Status: gap.StatusOpen}},
		CountResult: 1,
	}
	h := newTestHandler(repo)

	_, out, err := h.listGaps(context.Background(), nil, ListGapsInput{AgentID: "agent-1", Limit: 20, Offset: 0})
	if err != nil {
		t.Fatalf("listGaps: %v", err)
	}
	if out.Total != 1 || len(out.Gaps) != 1 {
		t.Fatalf("out = %+v, want exactly 1 gap", out)
	}
	if out.Gaps[0].ID != "gap-1" || out.Gaps[0].Category != "capability" {
		t.Errorf("gap = %+v, want id=gap-1 category=capability", out.Gaps[0])
	}
	if repo.LastListAgentID != "agent-1" {
		t.Errorf("List called with agentID=%q, want %q", repo.LastListAgentID, "agent-1")
	}
}

func TestListGapsTool_RepoError_ReturnsWrappedError(t *testing.T) {
	repo := &mock.GapRepository{CountErr: apperrors.NewAppError(apperrors.Internal, "boom")}
	h := newTestHandler(repo)

	_, _, err := h.listGaps(context.Background(), nil, ListGapsInput{AgentID: "agent-1"})
	if err == nil {
		t.Fatal("listGaps() = nil error, want the repo error wrapped")
	}
}

func TestListGapsTool_UnknownAgent_ReturnsError(t *testing.T) {
	repo := &mock.GapRepository{}
	agentRepo := &mock.AgentRepository{GetByIDErr: apperrors.NewAppError(apperrors.NotFound, "agent not found")}
	h := &Handler{
		gapApp: &app.GapApp{
			Queries: app.GapQueries{
				ListGaps: queries.NewListGapsQuery(repo, agentRepo),
			},
		},
	}

	_, _, err := h.listGaps(context.Background(), nil, ListGapsInput{AgentID: "missing-agent"})
	if err == nil {
		t.Fatal("listGaps() = nil error, want the agent-not-found error wrapped")
	}
}

func TestListGapsTool_StatusFilter_PassedToQuery(t *testing.T) {
	repo := &mock.GapRepository{}
	h := newTestHandler(repo)

	_, _, err := h.listGaps(context.Background(), nil, ListGapsInput{AgentID: "agent-1", Status: "open"})
	if err != nil {
		t.Fatalf("listGaps: %v", err)
	}
	if repo.LastListStatus != gap.StatusOpen {
		t.Errorf("List called with status=%q, want %q", repo.LastListStatus, gap.StatusOpen)
	}
}

func TestListGapsTool_InvalidStatus_ReturnsError(t *testing.T) {
	repo := &mock.GapRepository{}
	h := newTestHandler(repo)

	_, _, err := h.listGaps(context.Background(), nil, ListGapsInput{AgentID: "agent-1", Status: "not-a-status"})
	if err == nil {
		t.Fatal("listGaps() = nil error, want an InvalidInput error for an unrecognized status")
	}
}

func TestGetGapTool_HappyPath_IncludesReferences(t *testing.T) {
	repo := &mock.GapRepository{
		GetByIDResult: &gap.Gap{ID: "gap-1", AgentID: "agent-1", Category: gap.CategoryImprovement, Status: gap.StatusOpen},
		References:    []*gap.Reference{{GapID: "gap-1", SessionID: "session-1", MessageID: "message-1"}},
	}
	h := newTestHandler(repo)

	_, out, err := h.getGap(context.Background(), nil, GetGapInput{ID: "gap-1", AgentID: "agent-1"})
	if err != nil {
		t.Fatalf("getGap: %v", err)
	}
	if out.ID != "gap-1" || out.Category != "improvement" {
		t.Errorf("out = %+v, want id=gap-1 category=improvement", out)
	}
	if len(out.References) != 1 || out.References[0].SessionID != "session-1" || out.References[0].MessageID != "message-1" {
		t.Errorf("references = %+v, want the configured reference", out.References)
	}
}

func TestGetGapTool_NotFound_ReturnsError(t *testing.T) {
	repo := &mock.GapRepository{GetByIDErr: apperrors.NewAppError(apperrors.NotFound, "gap not found")}
	h := newTestHandler(repo)

	_, _, err := h.getGap(context.Background(), nil, GetGapInput{ID: "missing", AgentID: "agent-1"})
	if err == nil {
		t.Fatal("getGap() = nil error, want the repo's NotFound error wrapped")
	}
}

// A gap that exists but belongs to a different agent must error the same as
// one that doesn't exist at all — unlike the REST route, gaps_get has no
// separate transport-level scoping, so this is the only place it's enforced.
func TestGetGapTool_BelongsToDifferentAgent_ReturnsError(t *testing.T) {
	repo := &mock.GapRepository{
		GetByIDResult: &gap.Gap{ID: "gap-1", AgentID: "agent-1"},
	}
	h := newTestHandler(repo)

	_, _, err := h.getGap(context.Background(), nil, GetGapInput{ID: "gap-1", AgentID: "agent-2"})
	if err == nil {
		t.Fatal("getGap() = nil error, want a NotFound error for a gap owned by a different agent")
	}
}

func TestResolveGapTool_HappyPath_IncludesReferences(t *testing.T) {
	repo := &mock.GapRepository{
		GetByIDResult: &gap.Gap{ID: "gap-1", AgentID: "agent-1", Status: gap.StatusOpen},
		References:    []*gap.Reference{{GapID: "gap-1", SessionID: "session-1"}},
	}
	h := newTestHandler(repo)

	_, out, err := h.resolveGap(context.Background(), nil, ResolveGapInput{ID: "gap-1", AgentID: "agent-1"})
	if err != nil {
		t.Fatalf("resolveGap: %v", err)
	}
	if out.Status != "resolved" {
		t.Errorf("status = %q, want resolved", out.Status)
	}
	if len(out.References) != 1 || out.References[0].SessionID != "session-1" {
		t.Errorf("references = %+v, want the configured reference", out.References)
	}
	if repo.LastUpdate == nil || repo.LastUpdate.Status != gap.StatusResolved {
		t.Errorf("repo.Update called with %+v, want status resolved", repo.LastUpdate)
	}
}

func TestResolveGapTool_BelongsToDifferentAgent_ReturnsError(t *testing.T) {
	repo := &mock.GapRepository{
		GetByIDResult: &gap.Gap{ID: "gap-1", AgentID: "agent-1", Status: gap.StatusOpen},
	}
	h := newTestHandler(repo)

	_, _, err := h.resolveGap(context.Background(), nil, ResolveGapInput{ID: "gap-1", AgentID: "agent-2"})
	if err == nil {
		t.Fatal("resolveGap() = nil error, want a NotFound error for a gap owned by a different agent")
	}
	if repo.LastUpdate != nil {
		t.Error("repo.Update was called, want no write once the ownership guard rejects the request")
	}
}

func TestDismissGapTool_HappyPath(t *testing.T) {
	repo := &mock.GapRepository{
		GetByIDResult: &gap.Gap{ID: "gap-1", AgentID: "agent-1", Status: gap.StatusOpen},
	}
	h := newTestHandler(repo)

	_, out, err := h.dismissGap(context.Background(), nil, DismissGapInput{
		ID: "gap-1", AgentID: "agent-1",
		DismissalCategory: gap.DismissalDuplicate, DismissalReason: "already reported",
	})
	if err != nil {
		t.Fatalf("dismissGap: %v", err)
	}
	if out.Status != "dismissed" || out.DismissalCategory != gap.DismissalDuplicate {
		t.Errorf("out = %+v, want status=dismissed dismissalCategory=%q", out, gap.DismissalDuplicate)
	}
}

func TestDismissGapTool_InvalidCategory_ReturnsError(t *testing.T) {
	repo := &mock.GapRepository{
		GetByIDResult: &gap.Gap{ID: "gap-1", AgentID: "agent-1", Status: gap.StatusOpen},
	}
	h := newTestHandler(repo)

	_, _, err := h.dismissGap(context.Background(), nil, DismissGapInput{
		ID: "gap-1", AgentID: "agent-1", DismissalCategory: "not-a-real-category",
	})
	if err == nil {
		t.Fatal("dismissGap() = nil error, want an InvalidInput error for an unrecognized category")
	}
	if repo.LastUpdate != nil {
		t.Error("repo.Update was called, want no write for an invalid category")
	}
}

func TestDismissGapTool_BelongsToDifferentAgent_ReturnsError(t *testing.T) {
	repo := &mock.GapRepository{
		GetByIDResult: &gap.Gap{ID: "gap-1", AgentID: "agent-1", Status: gap.StatusOpen},
	}
	h := newTestHandler(repo)

	_, _, err := h.dismissGap(context.Background(), nil, DismissGapInput{
		ID: "gap-1", AgentID: "agent-2", DismissalCategory: gap.DismissalDuplicate,
	})
	if err == nil {
		t.Fatal("dismissGap() = nil error, want a NotFound error for a gap owned by a different agent")
	}
	if repo.LastUpdate != nil {
		t.Error("repo.Update was called, want no write once the ownership guard rejects the request")
	}
}

func TestReopenGapTool_HappyPath(t *testing.T) {
	repo := &mock.GapRepository{
		GetByIDResult: &gap.Gap{ID: "gap-1", AgentID: "agent-1", Status: gap.StatusResolved},
	}
	h := newTestHandler(repo)

	_, out, err := h.reopenGap(context.Background(), nil, ReopenGapInput{ID: "gap-1", AgentID: "agent-1"})
	if err != nil {
		t.Fatalf("reopenGap: %v", err)
	}
	if out.Status != "open" {
		t.Errorf("status = %q, want open", out.Status)
	}
}

func TestReopenGapTool_Terminal_ReturnsError(t *testing.T) {
	repo := &mock.GapRepository{
		GetByIDResult: &gap.Gap{ID: "gap-1", AgentID: "agent-1", Status: gap.StatusDismissed, DismissalCategory: gap.DismissalUnrelated},
	}
	h := newTestHandler(repo)

	_, _, err := h.reopenGap(context.Background(), nil, ReopenGapInput{ID: "gap-1", AgentID: "agent-1"})
	if err == nil {
		t.Fatal("reopenGap() = nil error, want a Conflict error for a gap dismissed as unrelated")
	}
}

func TestReopenGapTool_BelongsToDifferentAgent_ReturnsError(t *testing.T) {
	repo := &mock.GapRepository{
		GetByIDResult: &gap.Gap{ID: "gap-1", AgentID: "agent-1", Status: gap.StatusResolved},
	}
	h := newTestHandler(repo)

	_, _, err := h.reopenGap(context.Background(), nil, ReopenGapInput{ID: "gap-1", AgentID: "agent-2"})
	if err == nil {
		t.Fatal("reopenGap() = nil error, want a NotFound error for a gap owned by a different agent")
	}
	if repo.LastUpdate != nil {
		t.Error("repo.Update was called, want no write once the ownership guard rejects the request")
	}
}
