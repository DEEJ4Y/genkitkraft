package commands

import (
	"context"

	apperrors "github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/gap"
	gaprepo "github.com/DEEJ4Y/genkitkraft/internal/ports/gap_repo"
)

type ReopenGapParams struct {
	ID string
}

type ReopenGapResult struct {
	Gap *gap.Gap
}

// ReopenGapCommand moves a resolved or dismissed gap back to open. Used by
// both the UI and the dedup pipeline's merge path, so the terminal rule
// (gap.Gap.IsTerminal) is enforced here once for both callers.
type ReopenGapCommand struct {
	repo gaprepo.GapRepository
}

func NewReopenGapCommand(repo gaprepo.GapRepository) *ReopenGapCommand {
	return &ReopenGapCommand{repo: repo}
}

func (c *ReopenGapCommand) Execute(ctx context.Context, params ReopenGapParams) (ReopenGapResult, error) {
	g, err := c.repo.GetByID(ctx, params.ID)
	if err != nil {
		return ReopenGapResult{}, err
	}
	if g.IsTerminal() {
		return ReopenGapResult{}, apperrors.NewAppError(apperrors.Conflict, "gap dismissed as unrelated cannot be reopened")
	}

	g.Status = gap.StatusOpen
	g.DismissalCategory = ""
	g.DismissalReason = ""

	if err := c.repo.Update(ctx, g); err != nil {
		return ReopenGapResult{}, err
	}
	return ReopenGapResult{Gap: g}, nil
}
