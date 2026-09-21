package queries

import (
	"context"

	apperrors "github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	gaprepo "github.com/DEEJ4Y/genkitkraft/internal/ports/gap_repo"
)

type GetGapParams struct {
	ID      string
	AgentID string
}

type GetGapResult struct {
	GapWithReferences
}

type GetGapQuery struct {
	repo gaprepo.GapRepository
}

func NewGetGapQuery(repo gaprepo.GapRepository) *GetGapQuery {
	return &GetGapQuery{repo: repo}
}

func (q *GetGapQuery) Execute(ctx context.Context, params GetGapParams) (GetGapResult, error) {
	g, err := q.repo.GetByID(ctx, params.ID)
	if err != nil {
		return GetGapResult{}, err
	}
	if g.AgentID != params.AgentID {
		return GetGapResult{}, apperrors.NewAppError(apperrors.NotFound, "gap not found")
	}

	refs, err := q.repo.ListReferences(ctx, g.ID)
	if err != nil {
		return GetGapResult{}, err
	}

	return GetGapResult{GapWithReferences{Gap: g, References: refs}}, nil
}
