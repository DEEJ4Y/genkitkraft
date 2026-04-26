package app

import (
	"github.com/DEEJ4Y/genkitkraft/internal/app/commands"
	"github.com/DEEJ4Y/genkitkraft/internal/app/executors"
	"github.com/DEEJ4Y/genkitkraft/internal/app/queries"
)

type AgentToolCommands struct {
	UpdateTools executors.ExecutorWithReturn[commands.UpdateAgentToolsParams, commands.UpdateAgentToolsResult]
}

type AgentToolQueries struct {
	GetTools executors.ExecutorWithReturn[queries.GetAgentToolsParams, queries.GetAgentToolsResult]
}

type AgentToolApp struct {
	Commands AgentToolCommands
	Queries  AgentToolQueries
}
