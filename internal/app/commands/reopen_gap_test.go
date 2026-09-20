package commands_test

import (
	"context"
	"testing"

	"github.com/DEEJ4Y/genkitkraft/internal/app/commands"
	apperrors "github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/gap"
	"github.com/DEEJ4Y/genkitkraft/resources/test/mock"
)

func TestReopenGap_Terminal_ReturnsConflict(t *testing.T) {
	repo := &mock.GapRepository{
		GetByIDResult: &gap.Gap{ID: "gap-1", Status: gap.StatusDismissed, DismissalCategory: gap.DismissalUnrelated},
	}
	cmd := commands.NewReopenGapCommand(repo)

	_, err := cmd.Execute(context.Background(), commands.ReopenGapParams{ID: "gap-1"})

	appErr, ok := apperrors.IsAppError(err)
	if !ok {
		t.Fatalf("want *errors.AppError, got %T: %v", err, err)
	}
	if appErr.Code() != apperrors.Conflict {
		t.Errorf("error code = %v, want Conflict", appErr.Code())
	}
	if repo.LastUpdate != nil {
		t.Error("repo.Update was called, want no write once the terminal guard rejects the request")
	}
}

// A non-unrelated dismissal is not terminal — it must be reopenable, unlike
// the "unrelated" case above.
func TestReopenGap_NonTerminalDismissal_Succeeds(t *testing.T) {
	repo := &mock.GapRepository{
		GetByIDResult: &gap.Gap{
			ID: "gap-1", Status: gap.StatusDismissed,
			DismissalCategory: gap.DismissalDuplicate, DismissalReason: "matched another report",
		},
	}
	cmd := commands.NewReopenGapCommand(repo)

	result, err := cmd.Execute(context.Background(), commands.ReopenGapParams{ID: "gap-1"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if repo.LastUpdate == nil {
		t.Fatal("repo.Update was not called")
	}
	if repo.LastUpdate.Status != gap.StatusOpen {
		t.Errorf("status = %q, want %q", repo.LastUpdate.Status, gap.StatusOpen)
	}
	if repo.LastUpdate.DismissalCategory != "" || repo.LastUpdate.DismissalReason != "" {
		t.Errorf("dismissal fields not cleared: category=%q reason=%q", repo.LastUpdate.DismissalCategory, repo.LastUpdate.DismissalReason)
	}
	if result.Gap.Status != gap.StatusOpen {
		t.Errorf("result.Gap.Status = %q, want %q", result.Gap.Status, gap.StatusOpen)
	}
}

func TestReopenGap_Resolved_Succeeds(t *testing.T) {
	repo := &mock.GapRepository{GetByIDResult: &gap.Gap{ID: "gap-1", Status: gap.StatusResolved}}
	cmd := commands.NewReopenGapCommand(repo)

	if _, err := cmd.Execute(context.Background(), commands.ReopenGapParams{ID: "gap-1"}); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if repo.LastUpdate.Status != gap.StatusOpen {
		t.Errorf("status = %q, want %q", repo.LastUpdate.Status, gap.StatusOpen)
	}
}
