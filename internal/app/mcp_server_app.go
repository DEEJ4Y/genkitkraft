package app

import (
	"github.com/DEEJ4Y/genkitkraft/internal/app/commands"
	"github.com/DEEJ4Y/genkitkraft/internal/app/executors"
	"github.com/DEEJ4Y/genkitkraft/internal/app/queries"
)

// McpServerApp groups all MCP server management use cases.
type McpServerApp struct {
	Commands McpServerCommands
	Queries  McpServerQueries
}

type McpServerCommands struct {
	CreateMcpServer executors.ExecutorWithReturn[commands.CreateMcpServerParams, commands.CreateMcpServerResult]
	UpdateMcpServer executors.ExecutorWithReturn[commands.UpdateMcpServerParams, commands.UpdateMcpServerResult]
	DeleteMcpServer executors.Executor[commands.DeleteMcpServerParams]
}

type McpServerQueries struct {
	ListMcpServers executors.ExecutorWithReturn[queries.ListMcpServersParams, queries.ListMcpServersResult]
	GetMcpServer   executors.ExecutorWithReturn[queries.GetMcpServerParams, queries.GetMcpServerResult]
}
