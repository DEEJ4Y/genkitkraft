package queries

import (
	"context"

	"github.com/DEEJ4Y/genkitkraft/internal/domain/prompt"
	promptrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/prompt_repo"
)

type ListPromptsParams struct {
	Limit  int
	Offset int
}

type ListPromptsResult struct {
	Prompts []*prompt.Prompt
	Total   int
}

type ListPromptsQuery struct {
	repo promptrepo.PromptRepository
}

func NewListPromptsQuery(repo promptrepo.PromptRepository) *ListPromptsQuery {
	return &ListPromptsQuery{repo: repo}
}

func (q *ListPromptsQuery) Execute(ctx context.Context, params ListPromptsParams) (ListPromptsResult, error) {
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

	total, err := q.repo.Count(ctx)
	if err != nil {
		return ListPromptsResult{}, err
	}

	prompts, err := q.repo.List(ctx, limit, offset)
	if err != nil {
		return ListPromptsResult{}, err
	}

	return ListPromptsResult{
		Prompts: prompts,
		Total:   total,
	}, nil
}
