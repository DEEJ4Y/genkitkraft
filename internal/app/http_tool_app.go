package app

import (
	"github.com/DEEJ4Y/genkitkraft/internal/app/commands"
	"github.com/DEEJ4Y/genkitkraft/internal/app/executors"
	"github.com/DEEJ4Y/genkitkraft/internal/app/queries"
)

// HttpToolApp groups all HTTP tool management use cases.
type HttpToolApp struct {
	Commands HttpToolCommands
	Queries  HttpToolQueries
}

type HttpToolCommands struct {
	CreateHttpTool executors.ExecutorWithReturn[commands.CreateHttpToolParams, commands.CreateHttpToolResult]
	UpdateHttpTool executors.ExecutorWithReturn[commands.UpdateHttpToolParams, commands.UpdateHttpToolResult]
	DeleteHttpTool executors.Executor[commands.DeleteHttpToolParams]
}

type HttpToolQueries struct {
	ListHttpTools executors.ExecutorWithReturn[queries.ListHttpToolsParams, queries.ListHttpToolsResult]
	GetHttpTool   executors.ExecutorWithReturn[queries.GetHttpToolParams, queries.GetHttpToolResult]
}
