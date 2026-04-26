package commands

import (
	"context"

	"github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	httptool "github.com/DEEJ4Y/genkitkraft/internal/domain/http_tool"
	httptoolrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/http_tool_repo"
)

type CreateHttpToolParams struct {
	Name         string
	Description  string
	Method       string
	URL          string
	Headers      []httptool.HttpToolHeader
	BodyTemplate string
	InputSchema  string
}

type CreateHttpToolResult struct {
	HttpTool *httptool.HttpTool
}

type CreateHttpToolCommand struct {
	repo httptoolrepo.HttpToolRepository
}

func NewCreateHttpToolCommand(repo httptoolrepo.HttpToolRepository) *CreateHttpToolCommand {
	return &CreateHttpToolCommand{repo: repo}
}

func (c *CreateHttpToolCommand) Execute(ctx context.Context, params CreateHttpToolParams) (CreateHttpToolResult, error) {
	if params.Name == "" {
		return CreateHttpToolResult{}, errors.NewAppError(errors.InvalidInput, "name is required")
	}
	if params.URL == "" {
		return CreateHttpToolResult{}, errors.NewAppError(errors.InvalidInput, "url is required")
	}

	method := params.Method
	if method == "" {
		method = "GET"
	}

	headers := params.Headers
	if headers == nil {
		headers = []httptool.HttpToolHeader{}
	}

	inputSchema := params.InputSchema
	if inputSchema == "" {
		inputSchema = "{}"
	}

	t := &httptool.HttpTool{
		Name:         params.Name,
		Description:  params.Description,
		Method:       method,
		URL:          params.URL,
		Headers:      headers,
		BodyTemplate: params.BodyTemplate,
		InputSchema:  inputSchema,
	}

	if err := c.repo.Create(ctx, t); err != nil {
		return CreateHttpToolResult{}, err
	}

	return CreateHttpToolResult{HttpTool: t}, nil
}
