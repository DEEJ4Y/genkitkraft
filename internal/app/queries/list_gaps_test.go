package queries_test

import (
	"context"
	"testing"

	"github.com/DEEJ4Y/genkitkraft/internal/app/queries"
	apperrors "github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/agent"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/gap"
	"github.com/DEEJ4Y/genkitkraft/resources/test/mock"
)

func TestListGaps_PaginationClamping(t *testing.T) {
	tests := []struct {
		name       string
		limit      int
		offset     int
		wantLimit  int
		wantOffset int
	}{
		{"non-positive limit defaults to 20", 0, 0, 20, 0},
		{"negative limit defaults to 20", -5, 0, 20, 0},
		{"limit over 100 is capped", 500, 0, 100, 0},
		{"negative offset floors to 0", 10, -3, 10, 0},
		{"in-range values pass through unchanged", 10, 5, 10, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.GapRepository{}
			agentRepo := &mock.AgentRepository{GetByIDResult: &agent.Agent{ID: "agent-1"}}
			q := queries.NewListGapsQuery(repo, agentRepo)

			if _, err := q.Execute(context.Background(), queries.ListGapsParams{
				AgentID: "agent-1", Limit: tt.limit, Offset: tt.offset,
			}); err != nil {
				t.Fatalf("Execute: %v", err)
			}

			if repo.LastListLimit != tt.wantLimit {
				t.Errorf("List called with limit=%d, want %d", repo.LastListLimit, tt.wantLimit)
			}
			if repo.LastListOffset != tt.wantOffset {
				t.Errorf("List called with offset=%d, want %d", repo.LastListOffset, tt.wantOffset)
			}
			if repo.LastListAgentID != "agent-1" {
				t.Errorf("List called with agentID=%q, want %q", repo.LastListAgentID, "agent-1")
			}
		})
	}
}

func TestListGaps_AttachesReferencesPerGap(t *testing.T) {
	repo := &mock.GapRepository{
		ListGaps:    []*gap.Gap{{ID: "gap-1"}, {ID: "gap-2"}},
		CountResult: 2,
		References:  []*gap.Reference{{GapID: "gap-1", SessionID: "session-1"}},
	}
	agentRepo := &mock.AgentRepository{GetByIDResult: &agent.Agent{ID: "agent-1"}}
	q := queries.NewListGapsQuery(repo, agentRepo)

	result, err := q.Execute(context.Background(), queries.ListGapsParams{AgentID: "agent-1", Limit: 20, Offset: 0})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if result.Total != 2 {
		t.Errorf("Total = %d, want 2", result.Total)
	}
	if len(result.Gaps) != 2 {
		t.Fatalf("len(Gaps) = %d, want 2", len(result.Gaps))
	}
	// ListReferences is called once per returned gap; the mock records the
	// last gap ID it was asked about, which must be the second (last) one.
	if repo.LastListReferencesGapID != "gap-2" {
		t.Errorf("ListReferences last called with gapID=%q, want %q", repo.LastListReferencesGapID, "gap-2")
	}
	for i, gwr := range result.Gaps {
		if len(gwr.References) != 1 {
			t.Errorf("Gaps[%d].References = %v, want the configured reference", i, gwr.References)
		}
	}
}

func TestListGaps_CountError_Propagates(t *testing.T) {
	repo := &mock.GapRepository{CountErr: apperrors.NewAppError(apperrors.Internal, "count failed")}
	agentRepo := &mock.AgentRepository{GetByIDResult: &agent.Agent{ID: "agent-1"}}
	q := queries.NewListGapsQuery(repo, agentRepo)

	_, err := q.Execute(context.Background(), queries.ListGapsParams{AgentID: "agent-1"})
	if appErr, ok := apperrors.IsAppError(err); !ok || appErr.Code() != apperrors.Internal {
		t.Fatalf("Execute() = %v, want the Count error", err)
	}
}

func TestListGaps_ListError_Propagates(t *testing.T) {
	repo := &mock.GapRepository{ListErr: apperrors.NewAppError(apperrors.Internal, "list failed")}
	agentRepo := &mock.AgentRepository{GetByIDResult: &agent.Agent{ID: "agent-1"}}
	q := queries.NewListGapsQuery(repo, agentRepo)

	_, err := q.Execute(context.Background(), queries.ListGapsParams{AgentID: "agent-1"})
	if appErr, ok := apperrors.IsAppError(err); !ok || appErr.Code() != apperrors.Internal {
		t.Fatalf("Execute() = %v, want the List error", err)
	}
}

func TestListGaps_UnknownAgent_ReturnsNotFound(t *testing.T) {
	repo := &mock.GapRepository{}
	agentRepo := &mock.AgentRepository{GetByIDErr: apperrors.NewAppError(apperrors.NotFound, "agent not found")}
	q := queries.NewListGapsQuery(repo, agentRepo)

	_, err := q.Execute(context.Background(), queries.ListGapsParams{AgentID: "missing-agent"})
	if appErr, ok := apperrors.IsAppError(err); !ok || appErr.Code() != apperrors.NotFound {
		t.Fatalf("Execute() = %v, want a NotFound error", err)
	}
	if repo.LastCountAgentID != "" || repo.LastListAgentID != "" {
		t.Errorf("Count/List were called on the gap repo, want them skipped once the agent lookup fails")
	}
}

func TestListGaps_StatusFilter_PassedToRepo(t *testing.T) {
	repo := &mock.GapRepository{}
	agentRepo := &mock.AgentRepository{GetByIDResult: &agent.Agent{ID: "agent-1"}}
	q := queries.NewListGapsQuery(repo, agentRepo)

	if _, err := q.Execute(context.Background(), queries.ListGapsParams{AgentID: "agent-1", Status: gap.StatusResolved}); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if repo.LastListStatus != gap.StatusResolved {
		t.Errorf("List called with status=%q, want %q", repo.LastListStatus, gap.StatusResolved)
	}
	if repo.LastCountStatus != gap.StatusResolved {
		t.Errorf("Count called with status=%q, want %q", repo.LastCountStatus, gap.StatusResolved)
	}
}

func TestListGaps_InvalidStatus_ReturnsInvalidInput(t *testing.T) {
	repo := &mock.GapRepository{}
	agentRepo := &mock.AgentRepository{GetByIDResult: &agent.Agent{ID: "agent-1"}}
	q := queries.NewListGapsQuery(repo, agentRepo)

	_, err := q.Execute(context.Background(), queries.ListGapsParams{AgentID: "agent-1", Status: gap.Status("bogus")})
	if appErr, ok := apperrors.IsAppError(err); !ok || appErr.Code() != apperrors.InvalidInput {
		t.Fatalf("Execute() = %v, want an InvalidInput error", err)
	}
	if repo.LastCountAgentID != "" || repo.LastListAgentID != "" || agentRepo.LastGetByID != "" {
		t.Error("agent/gap repo were called, want status validated before any lookup")
	}
}
