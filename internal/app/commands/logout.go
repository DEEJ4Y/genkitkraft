package commands

import (
	"context"

	"github.com/DEEJ4Y/genkitkraft/internal/ports/session"
)

type LogoutParams struct {
	Token string
}

type LogoutCommand struct {
	sessionStore session.Store
}

func NewLogoutCommand(sessionStore session.Store) *LogoutCommand {
	return &LogoutCommand{sessionStore: sessionStore}
}

func (c *LogoutCommand) Execute(ctx context.Context, params LogoutParams) error {
	return c.sessionStore.Delete(ctx, params.Token)
}
