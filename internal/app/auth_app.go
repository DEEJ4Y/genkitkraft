package app

import (
	"github.com/DEEJ4Y/genkitkraft/internal/app/commands"
	"github.com/DEEJ4Y/genkitkraft/internal/app/executors"
	"github.com/DEEJ4Y/genkitkraft/internal/app/queries"
)

// AuthApp groups all authentication use cases.
type AuthApp struct {
	Commands AuthCommands
	Queries  AuthQueries
}

type AuthCommands struct {
	Login  executors.ExecutorWithReturn[commands.LoginParams, commands.LoginResult]
	Logout executors.Executor[commands.LogoutParams]
}

type AuthQueries struct {
	GetMe         executors.ExecutorWithReturn[queries.GetMeParams, queries.GetMeResult]
	GetAuthStatus executors.ExecutorWithReturn[queries.GetAuthStatusParams, queries.GetAuthStatusResult]
}
