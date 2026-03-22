package commands

import (
	"context"

	promptrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/prompt_repo"
)

type DeletePromptParams struct {
	ID string
}

type DeletePromptCommand struct {
	repo promptrepo.PromptRepository
}

func NewDeletePromptCommand(repo promptrepo.PromptRepository) *DeletePromptCommand {
	return &DeletePromptCommand{repo: repo}
}

func (c *DeletePromptCommand) Execute(ctx context.Context, params DeletePromptParams) error {
	return c.repo.Delete(ctx, params.ID)
}
