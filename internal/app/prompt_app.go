package app

import (
	"github.com/DEEJ4Y/genkitkraft/internal/app/commands"
	"github.com/DEEJ4Y/genkitkraft/internal/app/executors"
	"github.com/DEEJ4Y/genkitkraft/internal/app/queries"
)

// PromptApp groups all prompt management use cases.
type PromptApp struct {
	Commands PromptCommands
	Queries  PromptQueries
}

type PromptCommands struct {
	CreatePrompt executors.ExecutorWithReturn[commands.CreatePromptParams, commands.CreatePromptResult]
	UpdatePrompt executors.ExecutorWithReturn[commands.UpdatePromptParams, commands.UpdatePromptResult]
	DeletePrompt executors.Executor[commands.DeletePromptParams]
}

type PromptQueries struct {
	ListPrompts executors.ExecutorWithReturn[queries.ListPromptsParams, queries.ListPromptsResult]
	GetPrompt   executors.ExecutorWithReturn[queries.GetPromptParams, queries.GetPromptResult]
}
