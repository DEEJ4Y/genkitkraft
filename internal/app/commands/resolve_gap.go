package commands

import (
	"context"

	apperrors "github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/gap"
	gaprepo "github.com/DEEJ4Y/genkitkraft/internal/ports/gap_repo"
)

type ResolveGapParams struct {
	ID string
}

type ResolveGapResult struct {
	Gap *gap.Gap
}

type ResolveGapCommand struct {
	repo gaprepo.GapRepository
}

func NewResolveGapCommand(repo gaprepo.GapRepository) *ResolveGapCommand {
	return &ResolveGapCommand{repo: repo}
}

func (c *ResolveGapCommand) Execute(ctx context.Context, params ResolveGapParams) (ResolveGapResult, error) {
	g, err := c.repo.GetByID(ctx, params.ID)
	if err != nil {
		return ResolveGapResult{}, err
	}
	if g.IsTerminal() {
		return ResolveGapResult{}, apperrors.NewAppError(apperrors.Conflict, "gap dismissed as unrelated is terminal")
	}

	g.Status = gap.StatusResolved
	g.DismissalCategory = ""
	g.DismissalReason = ""

	if err := c.repo.Update(ctx, g); err != nil {
		return ResolveGapResult{}, err
	}
	return ResolveGapResult{Gap: g}, nil
}
