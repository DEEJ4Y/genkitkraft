package queries_test

import (
	"context"
	"testing"

	"github.com/DEEJ4Y/genkitkraft/internal/app/queries"
	apperrors "github.com/DEEJ4Y/genkitkraft/internal/common/errors"
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
			q := queries.NewListGapsQuery(repo)

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
	q := queries.NewListGapsQuery(repo)

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
	q := queries.NewListGapsQuery(repo)

	_, err := q.Execute(context.Background(), queries.ListGapsParams{AgentID: "agent-1"})
	if appErr, ok := apperrors.IsAppError(err); !ok || appErr.Code() != apperrors.Internal {
		t.Fatalf("Execute() = %v, want the Count error", err)
	}
}

func TestListGaps_ListError_Propagates(t *testing.T) {
	repo := &mock.GapRepository{ListErr: apperrors.NewAppError(apperrors.Internal, "list failed")}
	q := queries.NewListGapsQuery(repo)

	_, err := q.Execute(context.Background(), queries.ListGapsParams{AgentID: "agent-1"})
	if appErr, ok := apperrors.IsAppError(err); !ok || appErr.Code() != apperrors.Internal {
		t.Fatalf("Execute() = %v, want the List error", err)
	}
}
