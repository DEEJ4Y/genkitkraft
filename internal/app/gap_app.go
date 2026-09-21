package app

import (
	"github.com/DEEJ4Y/genkitkraft/internal/app/commands"
	"github.com/DEEJ4Y/genkitkraft/internal/app/executors"
	"github.com/DEEJ4Y/genkitkraft/internal/app/queries"
)

// GapApp groups all gap review use cases.
type GapApp struct {
	Commands GapCommands
	Queries  GapQueries
}

type GapCommands struct {
	DismissGap executors.ExecutorWithReturn[commands.DismissGapParams, commands.DismissGapResult]
	ResolveGap executors.ExecutorWithReturn[commands.ResolveGapParams, commands.ResolveGapResult]
	ReopenGap  executors.ExecutorWithReturn[commands.ReopenGapParams, commands.ReopenGapResult]
}

type GapQueries struct {
	ListGaps executors.ExecutorWithReturn[queries.ListGapsParams, queries.ListGapsResult]
	GetGap   executors.ExecutorWithReturn[queries.GetGapParams, queries.GetGapResult]
}
