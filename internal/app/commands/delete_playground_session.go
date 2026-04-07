package commands

import (
	"context"

	playgroundrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/playground_repo"
)

type DeletePlaygroundSessionParams struct {
	ID string
}

type DeletePlaygroundSessionCommand struct {
	repo playgroundrepo.PlaygroundRepository
}

func NewDeletePlaygroundSessionCommand(repo playgroundrepo.PlaygroundRepository) *DeletePlaygroundSessionCommand {
	return &DeletePlaygroundSessionCommand{repo: repo}
}

func (c *DeletePlaygroundSessionCommand) Execute(ctx context.Context, params DeletePlaygroundSessionParams) error {
	return c.repo.DeleteSession(ctx, params.ID)
}
