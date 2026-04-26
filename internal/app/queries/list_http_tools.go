package queries

import (
	"context"

	httptool "github.com/DEEJ4Y/genkitkraft/internal/domain/http_tool"
	httptoolrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/http_tool_repo"
)

type ListHttpToolsParams struct {
	Limit  int
	Offset int
}

type ListHttpToolsResult struct {
	HttpTools []*httptool.HttpTool
	Total     int
}

type ListHttpToolsQuery struct {
	repo httptoolrepo.HttpToolRepository
}

func NewListHttpToolsQuery(repo httptoolrepo.HttpToolRepository) *ListHttpToolsQuery {
	return &ListHttpToolsQuery{repo: repo}
}

func (q *ListHttpToolsQuery) Execute(ctx context.Context, params ListHttpToolsParams) (ListHttpToolsResult, error) {
	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	offset := params.Offset
	if offset < 0 {
		offset = 0
	}

	total, err := q.repo.Count(ctx)
	if err != nil {
		return ListHttpToolsResult{}, err
	}

	tools, err := q.repo.List(ctx, limit, offset)
	if err != nil {
		return ListHttpToolsResult{}, err
	}

	return ListHttpToolsResult{
		HttpTools: tools,
		Total:     total,
	}, nil
}
