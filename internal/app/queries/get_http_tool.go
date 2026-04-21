package queries

import (
	"context"

	httptool "github.com/DEEJ4Y/genkitkraft/internal/domain/http_tool"
	httptoolrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/http_tool_repo"
)

type GetHttpToolParams struct {
	ID string
}

type GetHttpToolResult struct {
	HttpTool *httptool.HttpTool
}

type GetHttpToolQuery struct {
	repo httptoolrepo.HttpToolRepository
}

func NewGetHttpToolQuery(repo httptoolrepo.HttpToolRepository) *GetHttpToolQuery {
	return &GetHttpToolQuery{repo: repo}
}

func (q *GetHttpToolQuery) Execute(ctx context.Context, params GetHttpToolParams) (GetHttpToolResult, error) {
	t, err := q.repo.GetByID(ctx, params.ID)
	if err != nil {
		return GetHttpToolResult{}, err
	}
	return GetHttpToolResult{HttpTool: t}, nil
}
