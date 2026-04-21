package commands

import (
	"context"

	httptoolrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/http_tool_repo"
)

type DeleteHttpToolParams struct {
	ID string
}

type DeleteHttpToolCommand struct {
	repo httptoolrepo.HttpToolRepository
}

func NewDeleteHttpToolCommand(repo httptoolrepo.HttpToolRepository) *DeleteHttpToolCommand {
	return &DeleteHttpToolCommand{repo: repo}
}

func (c *DeleteHttpToolCommand) Execute(ctx context.Context, params DeleteHttpToolParams) error {
	return c.repo.Delete(ctx, params.ID)
}
