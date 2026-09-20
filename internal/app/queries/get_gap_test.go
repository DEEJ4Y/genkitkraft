package queries_test

import (
	"context"
	"testing"

	"github.com/DEEJ4Y/genkitkraft/internal/app/queries"
	apperrors "github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/gap"
	"github.com/DEEJ4Y/genkitkraft/resources/test/mock"
)

func TestGetGap_HappyPath_IncludesReferences(t *testing.T) {
	repo := &mock.GapRepository{
		GetByIDResult: &gap.Gap{ID: "gap-1", AgentID: "agent-1", Context: "c", Details: "d"},
		References:    []*gap.Reference{{GapID: "gap-1", SessionID: "session-1", MessageID: "message-1"}},
	}
	q := queries.NewGetGapQuery(repo)

	result, err := q.Execute(context.Background(), queries.GetGapParams{ID: "gap-1"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if result.Gap.ID != "gap-1" {
		t.Errorf("Gap.ID = %q, want %q", result.Gap.ID, "gap-1")
	}
	if len(result.References) != 1 || result.References[0].MessageID != "message-1" {
		t.Errorf("References = %+v, want the configured reference", result.References)
	}
	if repo.LastGetByID != "gap-1" {
		t.Errorf("GetByID called with %q, want %q", repo.LastGetByID, "gap-1")
	}
}

func TestGetGap_NotFound_Propagates(t *testing.T) {
	repo := &mock.GapRepository{GetByIDErr: apperrors.NewAppError(apperrors.NotFound, "gap not found")}
	q := queries.NewGetGapQuery(repo)

	_, err := q.Execute(context.Background(), queries.GetGapParams{ID: "missing"})
	appErr, ok := apperrors.IsAppError(err)
	if !ok || appErr.Code() != apperrors.NotFound {
		t.Fatalf("Execute() = %v, want NotFound AppError", err)
	}
}
