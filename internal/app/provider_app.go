package app

import (
	"github.com/DEEJ4Y/genkitkraft/internal/app/commands"
	"github.com/DEEJ4Y/genkitkraft/internal/app/executors"
	"github.com/DEEJ4Y/genkitkraft/internal/app/queries"
)

// ProviderApp groups all provider management use cases.
type ProviderApp struct {
	Commands ProviderCommands
	Queries  ProviderQueries
}

type ProviderCommands struct {
	CreateProvider executors.ExecutorWithReturn[commands.CreateProviderParams, commands.CreateProviderResult]
	UpdateProvider executors.ExecutorWithReturn[commands.UpdateProviderParams, commands.UpdateProviderResult]
	DeleteProvider executors.Executor[commands.DeleteProviderParams]
	TestProvider   executors.ExecutorWithReturn[commands.TestProviderParams, commands.TestProviderResult]
}

type ProviderQueries struct {
	ListProviders     executors.ExecutorWithReturn[queries.ListProvidersParams, queries.ListProvidersResult]
	GetProvider       executors.ExecutorWithReturn[queries.GetProviderParams, queries.GetProviderResult]
	ListProviderTypes executors.ExecutorWithReturn[queries.ListProviderTypesParams, queries.ListProviderTypesResult]
}
