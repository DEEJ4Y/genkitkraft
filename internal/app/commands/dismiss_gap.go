package commands

import (
	"context"

	apperrors "github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/gap"
	gaprepo "github.com/DEEJ4Y/genkitkraft/internal/ports/gap_repo"
)

type DismissGapParams struct {
	ID                string
	AgentID           string
	DismissalCategory string
	DismissalReason   string
}

type DismissGapResult struct {
	Gap *gap.Gap
}

// DismissGapCommand marks a gap dismissed. If DismissalCategory is
// "unrelated", the gap becomes terminal — see gap.Gap.IsTerminal.
type DismissGapCommand struct {
	repo gaprepo.GapRepository
}

func NewDismissGapCommand(repo gaprepo.GapRepository) *DismissGapCommand {
	return &DismissGapCommand{repo: repo}
}

func (c *DismissGapCommand) Execute(ctx context.Context, params DismissGapParams) (DismissGapResult, error) {
	switch params.DismissalCategory {
	case gap.DismissalUnrelated, gap.DismissalInsufficientDetail, gap.DismissalDuplicate, gap.DismissalOther:
	default:
		return DismissGapResult{}, apperrors.NewAppError(apperrors.InvalidInput,
			"dismissal category must be one of: unrelated, insufficient_detail, duplicate, other")
	}

	g, err := c.repo.GetByID(ctx, params.ID)
	if err != nil {
		return DismissGapResult{}, err
	}
	if g.AgentID != params.AgentID {
		return DismissGapResult{}, apperrors.NewAppError(apperrors.NotFound, "gap not found")
	}
	if g.IsTerminal() {
		return DismissGapResult{}, apperrors.NewAppError(apperrors.Conflict, "gap dismissed as unrelated is terminal")
	}

	g.Status = gap.StatusDismissed
	g.DismissalCategory = params.DismissalCategory
	g.DismissalReason = params.DismissalReason

	if err := c.repo.Update(ctx, g); err != nil {
		return DismissGapResult{}, err
	}
	return DismissGapResult{Gap: g}, nil
}
