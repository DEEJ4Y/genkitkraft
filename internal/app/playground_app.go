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
	StartStream   executors.ExecutorWithReturn[commands.StartPlaygroundStreamParams, commands.StartPlaygroundStreamResult]
	CancelStream  executors.Executor[commands.CancelPlaygroundStreamParams]
	FailStream    executors.Executor[commands.FailPlaygroundStreamParams]
}

type PlaygroundQueries struct {
	ListSessions    executors.ExecutorWithReturn[queries.ListPlaygroundSessionsParams, queries.ListPlaygroundSessionsResult]
	GetSession      executors.ExecutorWithReturn[queries.GetPlaygroundSessionParams, queries.GetPlaygroundSessionResult]
	ListMessages    executors.ExecutorWithReturn[queries.ListPlaygroundMessagesParams, queries.ListPlaygroundMessagesResult]
	ResolveConfig   executors.ExecutorWithReturn[queries.ResolvePlaygroundConfigParams, queries.ResolvePlaygroundConfigResult]
	GetStreamChunks executors.ExecutorWithReturn[queries.GetPlaygroundStreamChunksParams, queries.GetPlaygroundStreamChunksResult]
}
