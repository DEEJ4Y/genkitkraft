package commands

import (
	"context"
	"time"

	apperrors "github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/gap"
	gaprepo "github.com/DEEJ4Y/genkitkraft/internal/ports/gap_repo"
)

type ReopenGapParams struct {
	ID      string
	AgentID string
}

type ReopenGapResult struct {
	Gap *gap.Gap
}

// ReopenGapCommand moves a resolved or dismissed gap back to open from the UI.
// The dedup pipeline's merge path uses the same domain method, gap.Gap.Reopen,
// so both callers record triage history the same way. The terminal rule
// (gap.Gap.IsTerminal) is enforced in both places.
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
	if g.AgentID != params.AgentID {
		return ReopenGapResult{}, apperrors.NewAppError(apperrors.NotFound, "gap not found")
	}
	if g.IsTerminal() {
		return ReopenGapResult{}, apperrors.NewAppError(apperrors.Conflict, "gap dismissed as unrelated cannot be reopened")
	}

	g.Reopen(time.Now().UTC())

	if err := c.repo.Update(ctx, g); err != nil {
		return ReopenGapResult{}, err
	}
	return ReopenGapResult{Gap: g}, nil
}
