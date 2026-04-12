package commands

import (
	"context"

	apperrors "github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	playgroundrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/playground_repo"
)

type DeletePlaygroundSessionParams struct {
	ID      string
	AgentID string
}

type DeletePlaygroundSessionCommand struct {
	repo playgroundrepo.PlaygroundRepository
}

func NewDeletePlaygroundSessionCommand(repo playgroundrepo.PlaygroundRepository) *DeletePlaygroundSessionCommand {
	return &DeletePlaygroundSessionCommand{repo: repo}
}

func (c *DeletePlaygroundSessionCommand) Execute(ctx context.Context, params DeletePlaygroundSessionParams) error {
	session, err := c.repo.GetSession(ctx, params.ID)
	if err != nil {
		return err
	}
	if session.AgentID != params.AgentID {
		return apperrors.NewAppError(apperrors.NotFound, "playground session not found")
	}
	return c.repo.DeleteSession(ctx, params.ID)
}
