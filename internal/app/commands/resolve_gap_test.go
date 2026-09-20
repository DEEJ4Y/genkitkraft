package commands_test

import (
	"context"
	"testing"

	"github.com/DEEJ4Y/genkitkraft/internal/app/commands"
	apperrors "github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/gap"
	"github.com/DEEJ4Y/genkitkraft/resources/test/mock"
)

func TestResolveGap_Terminal_ReturnsConflict(t *testing.T) {
	repo := &mock.GapRepository{
		GetByIDResult: &gap.Gap{ID: "gap-1", Status: gap.StatusDismissed, DismissalCategory: gap.DismissalUnrelated},
	}
	cmd := commands.NewResolveGapCommand(repo)

	_, err := cmd.Execute(context.Background(), commands.ResolveGapParams{ID: "gap-1"})

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

func TestResolveGap_HappyPath_ClearsAnyPriorDismissal(t *testing.T) {
	repo := &mock.GapRepository{
		GetByIDResult: &gap.Gap{
			ID: "gap-1", Status: gap.StatusDismissed,
			DismissalCategory: gap.DismissalDuplicate, DismissalReason: "seen before",
		},
	}
	cmd := commands.NewResolveGapCommand(repo)

	result, err := cmd.Execute(context.Background(), commands.ResolveGapParams{ID: "gap-1"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if repo.LastUpdate == nil {
		t.Fatal("repo.Update was not called")
	}
	if repo.LastUpdate.Status != gap.StatusResolved {
		t.Errorf("status = %q, want %q", repo.LastUpdate.Status, gap.StatusResolved)
	}
	if repo.LastUpdate.DismissalCategory != "" || repo.LastUpdate.DismissalReason != "" {
		t.Errorf("dismissal fields not cleared: category=%q reason=%q", repo.LastUpdate.DismissalCategory, repo.LastUpdate.DismissalReason)
	}
	if result.Gap.Status != gap.StatusResolved {
		t.Errorf("result.Gap.Status = %q, want %q", result.Gap.Status, gap.StatusResolved)
	}
}

func TestResolveGap_NotFound_PropagatesRepoError(t *testing.T) {
	repo := &mock.GapRepository{GetByIDErr: apperrors.NewAppError(apperrors.NotFound, "gap not found")}
	cmd := commands.NewResolveGapCommand(repo)

	_, err := cmd.Execute(context.Background(), commands.ResolveGapParams{ID: "missing"})

	appErr, ok := apperrors.IsAppError(err)
	if !ok || appErr.Code() != apperrors.NotFound {
		t.Fatalf("want NotFound AppError, got %v", err)
	}
}
