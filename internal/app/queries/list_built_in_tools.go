package queries

import (
	"context"

	builtintool "github.com/DEEJ4Y/genkitkraft/internal/domain/builtin_tool"
)

type ListBuiltInToolsResult struct {
	BuiltInTools []builtintool.BuiltInTool
}

type ListBuiltInToolsQuery struct{}

func NewListBuiltInToolsQuery() *ListBuiltInToolsQuery {
	return &ListBuiltInToolsQuery{}
}

func (q *ListBuiltInToolsQuery) Execute(_ context.Context) (ListBuiltInToolsResult, error) {
	return ListBuiltInToolsResult{BuiltInTools: builtintool.Registry}, nil
}
