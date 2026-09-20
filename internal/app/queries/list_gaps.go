package queries

import (
	"context"

	"github.com/DEEJ4Y/genkitkraft/internal/domain/gap"
	agentrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/agent_repo"
	gaprepo "github.com/DEEJ4Y/genkitkraft/internal/ports/gap_repo"
)

type ListGapsParams struct {
	AgentID string
	Limit   int
	Offset  int
}

// GapWithReferences pairs a Gap with the conversations it was observed in.
type GapWithReferences struct {
	Gap        *gap.Gap
	References []*gap.Reference
}

type ListGapsResult struct {
	Gaps  []GapWithReferences
	Total int
}

type ListGapsQuery struct {
	repo      gaprepo.GapRepository
	agentRepo agentrepo.AgentRepository
}

func NewListGapsQuery(repo gaprepo.GapRepository, agentRepo agentrepo.AgentRepository) *ListGapsQuery {
	return &ListGapsQuery{repo: repo, agentRepo: agentRepo}
}

func (q *ListGapsQuery) Execute(ctx context.Context, params ListGapsParams) (ListGapsResult, error) {
	if _, err := q.agentRepo.GetByID(ctx, params.AgentID); err != nil {
		return ListGapsResult{}, err
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	offset := params.Offset
	if offset < 0 {
		offset = 0
	}

	total, err := q.repo.Count(ctx, params.AgentID)
	if err != nil {
		return ListGapsResult{}, err
	}

	gaps, err := q.repo.List(ctx, params.AgentID, limit, offset)
	if err != nil {
		return ListGapsResult{}, err
	}

	result := make([]GapWithReferences, 0, len(gaps))
	for _, g := range gaps {
		refs, err := q.repo.ListReferences(ctx, g.ID)
		if err != nil {
			return ListGapsResult{}, err
		}
		result = append(result, GapWithReferences{Gap: g, References: refs})
	}

	return ListGapsResult{Gaps: result, Total: total}, nil
}
