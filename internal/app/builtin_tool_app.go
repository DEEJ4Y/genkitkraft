package app

import (
	"github.com/DEEJ4Y/genkitkraft/internal/app/queries"
)

// BuiltInToolApp groups all built-in tool use cases.
type BuiltInToolApp struct {
	Queries BuiltInToolQueries
}

type BuiltInToolQueries struct {
	ListBuiltInTools *queries.ListBuiltInToolsQuery
}
