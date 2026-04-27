package mcphandler

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/DEEJ4Y/genkitkraft/internal/app/commands"
	"github.com/DEEJ4Y/genkitkraft/internal/app/queries"
	httptool "github.com/DEEJ4Y/genkitkraft/internal/domain/http_tool"
)

// --- Input/Output types ---

type ListHttpToolsInput struct {
	Limit  int `json:"limit" jsonschema:"max number of HTTP tools to return (default 20, max 100)"`
	Offset int `json:"offset" jsonschema:"offset for pagination"`
}

type HttpToolHeaderOutput struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type HttpToolOutput struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Method       string                 `json:"method"`
	URL          string                 `json:"url"`
	Headers      []HttpToolHeaderOutput `json:"headers,omitempty"`
	BodyTemplate string                 `json:"body_template,omitempty"`
	InputSchema  string                 `json:"input_schema,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

type ListHttpToolsOutput struct {
	HttpTools []HttpToolOutput `json:"http_tools"`
	Total     int              `json:"total"`
}

type GetHttpToolInput struct {
	ID string `json:"id" jsonschema:"HTTP tool ID"`
}

type HttpToolHeaderInput struct {
	Name  string `json:"name" jsonschema:"header name"`
	Value string `json:"value" jsonschema:"header value"`
}

type CreateHttpToolInput struct {
	Name         string                `json:"name" jsonschema:"tool name (required)"`
	Description  string                `json:"description" jsonschema:"tool description (required)"`
	Method       string                `json:"method" jsonschema:"HTTP method: GET, POST, PUT, DELETE, PATCH (required)"`
	URL          string                `json:"url" jsonschema:"URL template (required)"`
	Headers      []HttpToolHeaderInput `json:"headers,omitempty" jsonschema:"request headers"`
	BodyTemplate string                `json:"body_template,omitempty" jsonschema:"request body template with {{placeholders}}"`
	InputSchema  string                `json:"input_schema,omitempty" jsonschema:"JSON schema for tool input parameters"`
}

type UpdateHttpToolInput struct {
	ID           string                 `json:"id" jsonschema:"HTTP tool ID (required)"`
	Name         *string                `json:"name,omitempty" jsonschema:"new tool name"`
	Description  *string                `json:"description,omitempty" jsonschema:"new description"`
	Method       *string                `json:"method,omitempty" jsonschema:"new HTTP method"`
	URL          *string                `json:"url,omitempty" jsonschema:"new URL template"`
	Headers      *[]HttpToolHeaderInput `json:"headers,omitempty" jsonschema:"new headers"`
	BodyTemplate *string                `json:"body_template,omitempty" jsonschema:"new body template"`
	InputSchema  *string                `json:"input_schema,omitempty" jsonschema:"new input schema"`
}

type DeleteHttpToolInput struct {
	ID string `json:"id" jsonschema:"HTTP tool ID to delete"`
}

// --- Tool registration ---

func (h *Handler) registerHttpToolTools(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "http_tools_list",
		Description: "List all HTTP tools with pagination.",
	}, h.listHttpTools)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "http_tools_get",
		Description: "Get an HTTP tool by ID.",
	}, h.getHttpTool)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "http_tools_create",
		Description: "Create a new HTTP tool (custom API call).",
	}, h.createHttpTool)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "http_tools_update",
		Description: "Update an existing HTTP tool. Only provided fields are changed.",
	}, h.updateHttpTool)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "http_tools_delete",
		Description: "Delete an HTTP tool by ID.",
	}, h.deleteHttpTool)
}

// --- Tool handlers ---

func (h *Handler) listHttpTools(ctx context.Context, _ *mcp.CallToolRequest, input ListHttpToolsInput) (*mcp.CallToolResult, ListHttpToolsOutput, error) {
	result, err := h.httpToolApp.Queries.ListHttpTools.Execute(ctx, queries.ListHttpToolsParams{
		Limit:  input.Limit,
		Offset: input.Offset,
	})
	if err != nil {
		return nil, ListHttpToolsOutput{}, fmt.Errorf("list http tools failed: %w", err)
	}
	tools := make([]HttpToolOutput, len(result.HttpTools))
	for i, t := range result.HttpTools {
		tools[i] = toHttpToolOutput(t)
	}
	return nil, ListHttpToolsOutput{HttpTools: tools, Total: result.Total}, nil
}

func (h *Handler) getHttpTool(ctx context.Context, _ *mcp.CallToolRequest, input GetHttpToolInput) (*mcp.CallToolResult, HttpToolOutput, error) {
	result, err := h.httpToolApp.Queries.GetHttpTool.Execute(ctx, queries.GetHttpToolParams{ID: input.ID})
	if err != nil {
		return nil, HttpToolOutput{}, fmt.Errorf("get http tool failed: %w", err)
	}
	return nil, toHttpToolOutput(result.HttpTool), nil
}

func (h *Handler) createHttpTool(ctx context.Context, _ *mcp.CallToolRequest, input CreateHttpToolInput) (*mcp.CallToolResult, HttpToolOutput, error) {
	headers := make([]httptool.HttpToolHeader, len(input.Headers))
	for i, hdr := range input.Headers {
		headers[i] = httptool.HttpToolHeader{Name: hdr.Name, Value: hdr.Value}
	}
	result, err := h.httpToolApp.Commands.CreateHttpTool.Execute(ctx, commands.CreateHttpToolParams{
		Name:         input.Name,
		Description:  input.Description,
		Method:       input.Method,
		URL:          input.URL,
		Headers:      headers,
		BodyTemplate: input.BodyTemplate,
		InputSchema:  input.InputSchema,
	})
	if err != nil {
		return nil, HttpToolOutput{}, fmt.Errorf("create http tool failed: %w", err)
	}
	return nil, toHttpToolOutput(result.HttpTool), nil
}

func (h *Handler) updateHttpTool(ctx context.Context, _ *mcp.CallToolRequest, input UpdateHttpToolInput) (*mcp.CallToolResult, HttpToolOutput, error) {
	params := commands.UpdateHttpToolParams{
		ID:           input.ID,
		Name:         input.Name,
		Description:  input.Description,
		Method:       input.Method,
		URL:          input.URL,
		BodyTemplate: input.BodyTemplate,
		InputSchema:  input.InputSchema,
	}
	if input.Headers != nil {
		headers := make([]httptool.HttpToolHeader, len(*input.Headers))
		for i, hdr := range *input.Headers {
			headers[i] = httptool.HttpToolHeader{Name: hdr.Name, Value: hdr.Value}
		}
		params.Headers = &headers
	}
	result, err := h.httpToolApp.Commands.UpdateHttpTool.Execute(ctx, params)
	if err != nil {
		return nil, HttpToolOutput{}, fmt.Errorf("update http tool failed: %w", err)
	}
	return nil, toHttpToolOutput(result.HttpTool), nil
}

func (h *Handler) deleteHttpTool(ctx context.Context, _ *mcp.CallToolRequest, input DeleteHttpToolInput) (*mcp.CallToolResult, any, error) {
	err := h.httpToolApp.Commands.DeleteHttpTool.Execute(ctx, commands.DeleteHttpToolParams{ID: input.ID})
	if err != nil {
		return nil, nil, fmt.Errorf("delete http tool failed: %w", err)
	}
	return nil, map[string]string{"status": "deleted"}, nil
}

func toHttpToolOutput(t *httptool.HttpTool) HttpToolOutput {
	headers := make([]HttpToolHeaderOutput, len(t.Headers))
	for i, hdr := range t.Headers {
		headers[i] = HttpToolHeaderOutput{Name: hdr.Name, Value: hdr.Value}
	}
	return HttpToolOutput{
		ID:           t.ID,
		Name:         t.Name,
		Description:  t.Description,
		Method:       t.Method,
		URL:          t.URL,
		Headers:      headers,
		BodyTemplate: t.BodyTemplate,
		InputSchema:  t.InputSchema,
		CreatedAt:    t.CreatedAt,
		UpdatedAt:    t.UpdatedAt,
	}
}
