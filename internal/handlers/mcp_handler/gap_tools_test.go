package mcphandler

import (
	"context"
	"testing"

	"github.com/DEEJ4Y/genkitkraft/internal/app"
	"github.com/DEEJ4Y/genkitkraft/internal/app/queries"
	apperrors "github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/gap"
	"github.com/DEEJ4Y/genkitkraft/resources/test/mock"
)

func newTestHandler(repo *mock.GapRepository) *Handler {
	return &Handler{
		gapApp: &app.GapApp{
			Queries: app.GapQueries{
				ListGaps: queries.NewListGapsQuery(repo),
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

func TestGetGapTool_HappyPath_IncludesReferences(t *testing.T) {
	repo := &mock.GapRepository{
		GetByIDResult: &gap.Gap{ID: "gap-1", AgentID: "agent-1", Category: gap.CategoryImprovement, Status: gap.StatusOpen},
		References:    []*gap.Reference{{GapID: "gap-1", SessionID: "session-1", MessageID: "message-1"}},
	}
	h := newTestHandler(repo)

	_, out, err := h.getGap(context.Background(), nil, GetGapInput{ID: "gap-1"})
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

	_, _, err := h.getGap(context.Background(), nil, GetGapInput{ID: "missing"})
	if err == nil {
		t.Fatal("getGap() = nil error, want the repo's NotFound error wrapped")
	}
}
