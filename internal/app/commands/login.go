package commands

import (
	"context"

	"github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	domainauth "github.com/DEEJ4Y/genkitkraft/internal/domain/auth"
	"github.com/DEEJ4Y/genkitkraft/internal/ports/hasher"
	"github.com/DEEJ4Y/genkitkraft/internal/ports/session"
)

type LoginParams struct {
	Username string
	Password string
	ClientIP string
}

type LoginResult struct {
	Token    string
	Username string
}

type LoginCommand struct {
	users          map[string]*domainauth.User
	sessionStore   session.Store
	passwordHasher hasher.PasswordHasher
}

func NewLoginCommand(
	users map[string]*domainauth.User,
	sessionStore session.Store,
	passwordHasher hasher.PasswordHasher,
) *LoginCommand {
	return &LoginCommand{
		users:          users,
		sessionStore:   sessionStore,
		passwordHasher: passwordHasher,
	}
}

func (c *LoginCommand) Execute(ctx context.Context, params LoginParams) (LoginResult, error) {
	user, ok := c.users[params.Username]
	if !ok {
		return LoginResult{}, errors.NewAppError(errors.Unauthorized, "invalid credentials")
	}

	if err := c.passwordHasher.Compare(user.PasswordHash, params.Password); err != nil {
		return LoginResult{}, errors.NewAppError(errors.Unauthorized, "invalid credentials")
	}

	token, err := c.sessionStore.Create(ctx, params.Username)
	if err != nil {
		return LoginResult{}, errors.NewAppError(errors.Internal, "failed to create session")
	}

	return LoginResult{
		Token:    token,
		Username: params.Username,
	}, nil
}
