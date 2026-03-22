package queries

import (
	"context"

	"github.com/DEEJ4Y/genkitkraft/internal/domain/prompt"
	promptrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/prompt_repo"
)

type GetPromptParams struct {
	ID string
}

type GetPromptResult struct {
	Prompt *prompt.Prompt
}

type GetPromptQuery struct {
	repo promptrepo.PromptRepository
}

func NewGetPromptQuery(repo promptrepo.PromptRepository) *GetPromptQuery {
	return &GetPromptQuery{repo: repo}
}

func (q *GetPromptQuery) Execute(ctx context.Context, params GetPromptParams) (GetPromptResult, error) {
	p, err := q.repo.GetByID(ctx, params.ID)
	if err != nil {
		return GetPromptResult{}, err
	}
	return GetPromptResult{Prompt: p}, nil
}
