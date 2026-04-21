package commands

import (
	"context"

	httptool "github.com/DEEJ4Y/genkitkraft/internal/domain/http_tool"
	httptoolrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/http_tool_repo"
)

type UpdateHttpToolParams struct {
	ID           string
	Name         *string
	Description  *string
	Method       *string
	URL          *string
	Headers      *[]httptool.HttpToolHeader
	BodyTemplate *string
	InputSchema  *string
}

type UpdateHttpToolResult struct {
	HttpTool *httptool.HttpTool
}

type UpdateHttpToolCommand struct {
	repo httptoolrepo.HttpToolRepository
}

func NewUpdateHttpToolCommand(repo httptoolrepo.HttpToolRepository) *UpdateHttpToolCommand {
	return &UpdateHttpToolCommand{repo: repo}
}

func (c *UpdateHttpToolCommand) Execute(ctx context.Context, params UpdateHttpToolParams) (UpdateHttpToolResult, error) {
	t, err := c.repo.GetByID(ctx, params.ID)
	if err != nil {
		return UpdateHttpToolResult{}, err
	}

	if params.Name != nil {
		t.Name = *params.Name
	}
	if params.Description != nil {
		t.Description = *params.Description
	}
	if params.Method != nil {
		t.Method = *params.Method
	}
	if params.URL != nil {
		t.URL = *params.URL
	}
	if params.Headers != nil {
		t.Headers = *params.Headers
	}
	if params.BodyTemplate != nil {
		t.BodyTemplate = *params.BodyTemplate
	}
	if params.InputSchema != nil {
		t.InputSchema = *params.InputSchema
	}

	if err := c.repo.Update(ctx, t); err != nil {
		return UpdateHttpToolResult{}, err
	}

	return UpdateHttpToolResult{HttpTool: t}, nil
}
