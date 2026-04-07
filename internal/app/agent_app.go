package app

import (
	"github.com/DEEJ4Y/genkitkraft/internal/app/commands"
	"github.com/DEEJ4Y/genkitkraft/internal/app/executors"
	"github.com/DEEJ4Y/genkitkraft/internal/app/queries"
)

// AgentApp groups all agent management use cases.
type AgentApp struct {
	Commands AgentCommands
	Queries  AgentQueries
}

type AgentCommands struct {
	CreateAgent executors.ExecutorWithReturn[commands.CreateAgentParams, commands.CreateAgentResult]
	UpdateAgent executors.ExecutorWithReturn[commands.UpdateAgentParams, commands.UpdateAgentResult]
	DeleteAgent executors.Executor[commands.DeleteAgentParams]
}

type AgentQueries struct {
	ListAgents executors.ExecutorWithReturn[queries.ListAgentsParams, queries.ListAgentsResult]
	GetAgent   executors.ExecutorWithReturn[queries.GetAgentParams, queries.GetAgentResult]
}
