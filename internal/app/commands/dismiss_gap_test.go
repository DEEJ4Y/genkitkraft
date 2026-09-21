package commands_test

import (
	"context"
	"testing"

	"github.com/DEEJ4Y/genkitkraft/internal/app/commands"
	apperrors "github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/gap"
	"github.com/DEEJ4Y/genkitkraft/resources/test/mock"
)

func TestDismissGap_InvalidCategory_ReturnsInvalidInputWithoutTouchingRepo(t *testing.T) {
	repo := &mock.GapRepository{}
	cmd := commands.NewDismissGapCommand(repo)

	_, err := cmd.Execute(context.Background(), commands.DismissGapParams{
		ID:                "gap-1",
		DismissalCategory: "not-a-real-category",
	})

	appErr, ok := apperrors.IsAppError(err)
	if !ok {
		t.Fatalf("want *errors.AppError, got %T: %v", err, err)
	}
	if appErr.Code() != apperrors.InvalidInput {
		t.Errorf("error code = %v, want InvalidInput", appErr.Code())
	}
	if repo.LastGetByID != "" {
		t.Errorf("repo.GetByID called with %q, want no call at all — category is validated first", repo.LastGetByID)
	}
	if repo.LastUpdate != nil {
		t.Error("repo.Update was called, want no write for an invalid category")
	}
}

// Dismissing a gap that is already terminal (dismissed as unrelated) must be
// rejected — otherwise a caller could "re-dismiss" it with a different,
// reopenable category and defeat the terminal rule.
func TestDismissGap_Terminal_ReturnsConflict(t *testing.T) {
	repo := &mock.GapRepository{
		GetByIDResult: &gap.Gap{ID: "gap-1", Status: gap.StatusDismissed, DismissalCategory: gap.DismissalUnrelated},
	}
	cmd := commands.NewDismissGapCommand(repo)

	_, err := cmd.Execute(context.Background(), commands.DismissGapParams{
		ID:                "gap-1",
		DismissalCategory: gap.DismissalDuplicate,
	})

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

func TestDismissGap_BelongsToDifferentAgent_ReturnsNotFound(t *testing.T) {
	repo := &mock.GapRepository{
		GetByIDResult: &gap.Gap{ID: "gap-1", AgentID: "agent-1", Status: gap.StatusOpen},
	}
	cmd := commands.NewDismissGapCommand(repo)

	_, err := cmd.Execute(context.Background(), commands.DismissGapParams{
		ID:                "gap-1",
		AgentID:           "agent-2",
		DismissalCategory: gap.DismissalDuplicate,
	})

	appErr, ok := apperrors.IsAppError(err)
	if !ok || appErr.Code() != apperrors.NotFound {
		t.Fatalf("Execute() = %v, want NotFound AppError", err)
	}
	if repo.LastUpdate != nil {
		t.Error("repo.Update was called, want no write once the ownership guard rejects the request")
	}
}

func TestDismissGap_HappyPath(t *testing.T) {
	repo := &mock.GapRepository{
		GetByIDResult: &gap.Gap{ID: "gap-1", Status: gap.StatusOpen},
	}
	cmd := commands.NewDismissGapCommand(repo)

	result, err := cmd.Execute(context.Background(), commands.DismissGapParams{
		ID:                "gap-1",
		DismissalCategory: gap.DismissalInsufficientDetail,
		DismissalReason:   "not enough context to act on",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if repo.LastUpdate == nil {
		t.Fatal("repo.Update was not called")
	}
	if repo.LastUpdate.Status != gap.StatusDismissed {
		t.Errorf("status = %q, want %q", repo.LastUpdate.Status, gap.StatusDismissed)
	}
	if repo.LastUpdate.DismissalCategory != gap.DismissalInsufficientDetail {
		t.Errorf("dismissalCategory = %q, want %q", repo.LastUpdate.DismissalCategory, gap.DismissalInsufficientDetail)
	}
	if repo.LastUpdate.DismissalReason != "not enough context to act on" {
		t.Errorf("dismissalReason = %q, want the given reason", repo.LastUpdate.DismissalReason)
	}
	if result.Gap.Status != gap.StatusDismissed {
		t.Errorf("result.Gap.Status = %q, want %q", result.Gap.Status, gap.StatusDismissed)
	}
}
