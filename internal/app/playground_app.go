package app

import (
	"github.com/DEEJ4Y/genkitkraft/internal/app/commands"
	"github.com/DEEJ4Y/genkitkraft/internal/app/executors"
	"github.com/DEEJ4Y/genkitkraft/internal/app/queries"
)

// PlaygroundApp groups all playground use cases.
type PlaygroundApp struct {
	Commands PlaygroundCommands
	Queries  PlaygroundQueries
}

type PlaygroundCommands struct {
	CreateSession executors.ExecutorWithReturn[commands.CreatePlaygroundSessionParams, commands.CreatePlaygroundSessionResult]
	DeleteSession executors.Executor[commands.DeletePlaygroundSessionParams]
	SaveMessage   executors.ExecutorWithReturn[commands.SavePlaygroundMessageParams, commands.SavePlaygroundMessageResult]
}

type PlaygroundQueries struct {
	ListSessions  executors.ExecutorWithReturn[queries.ListPlaygroundSessionsParams, queries.ListPlaygroundSessionsResult]
	ListMessages  executors.ExecutorWithReturn[queries.ListPlaygroundMessagesParams, queries.ListPlaygroundMessagesResult]
	ResolveConfig executors.ExecutorWithReturn[queries.ResolvePlaygroundConfigParams, queries.ResolvePlaygroundConfigResult]
}
