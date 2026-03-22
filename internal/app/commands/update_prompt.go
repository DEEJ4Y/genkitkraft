package commands

import (
	"context"

	"github.com/DEEJ4Y/genkitkraft/internal/domain/prompt"
	promptrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/prompt_repo"
)

type UpdatePromptParams struct {
	ID      string
	Name    *string
	Content *string
}

type UpdatePromptResult struct {
	Prompt *prompt.Prompt
}

type UpdatePromptCommand struct {
	repo promptrepo.PromptRepository
}

func NewUpdatePromptCommand(repo promptrepo.PromptRepository) *UpdatePromptCommand {
	return &UpdatePromptCommand{repo: repo}
}

func (c *UpdatePromptCommand) Execute(ctx context.Context, params UpdatePromptParams) (UpdatePromptResult, error) {
	p, err := c.repo.GetByID(ctx, params.ID)
	if err != nil {
		return UpdatePromptResult{}, err
	}

	if params.Name != nil {
		p.Name = *params.Name
	}
	if params.Content != nil {
		p.Content = *params.Content
	}

	if err := c.repo.Update(ctx, p); err != nil {
		return UpdatePromptResult{}, err
	}

	return UpdatePromptResult{Prompt: p}, nil
}
