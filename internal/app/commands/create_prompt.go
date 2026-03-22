package commands

import (
	"context"

	"github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/prompt"
	promptrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/prompt_repo"
)

type CreatePromptParams struct {
	Name    string
	Content string
}

type CreatePromptResult struct {
	Prompt *prompt.Prompt
}

type CreatePromptCommand struct {
	repo promptrepo.PromptRepository
}

func NewCreatePromptCommand(repo promptrepo.PromptRepository) *CreatePromptCommand {
	return &CreatePromptCommand{repo: repo}
}

func (c *CreatePromptCommand) Execute(ctx context.Context, params CreatePromptParams) (CreatePromptResult, error) {
	if params.Name == "" {
		return CreatePromptResult{}, errors.NewAppError(errors.InvalidInput, "name is required")
	}

	p := &prompt.Prompt{
		Name:    params.Name,
		Content: params.Content,
	}

	if err := c.repo.Create(ctx, p); err != nil {
		return CreatePromptResult{}, err
	}

	return CreatePromptResult{Prompt: p}, nil
}
