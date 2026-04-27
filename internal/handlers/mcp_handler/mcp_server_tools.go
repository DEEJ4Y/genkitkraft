package mcphandler

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/DEEJ4Y/genkitkraft/internal/app/commands"
	"github.com/DEEJ4Y/genkitkraft/internal/app/queries"
	mcpserver "github.com/DEEJ4Y/genkitkraft/internal/domain/mcp_server"
)

// --- Input/Output types ---

type ListMcpServersInput struct {
	Limit  int `json:"limit" jsonschema:"max number of MCP servers to return (default 20, max 100)"`
	Offset int `json:"offset" jsonschema:"offset for pagination"`
}

type McpServerHeaderOutput struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type McpServerOutput struct {
	ID        string                  `json:"id"`
	Name      string                  `json:"name"`
	Transport string                  `json:"transport"`
	URL       string                  `json:"url"`
	Headers   []McpServerHeaderOutput `json:"headers,omitempty"`
	CreatedAt time.Time               `json:"created_at"`
	UpdatedAt time.Time               `json:"updated_at"`
}

type ListMcpServersOutput struct {
	McpServers []McpServerOutput `json:"mcp_servers"`
	Total      int               `json:"total"`
}

type GetMcpServerInput struct {
	ID string `json:"id" jsonschema:"MCP server ID"`
}

type McpServerHeaderInput struct {
	Name  string `json:"name" jsonschema:"header name"`
	Value string `json:"value" jsonschema:"header value"`
}

type CreateMcpServerInput struct {
	Name      string                 `json:"name" jsonschema:"server name (required)"`
	Transport string                 `json:"transport" jsonschema:"transport type: sse or streamableHttp (required)"`
	URL       string                 `json:"url" jsonschema:"server URL (required)"`
	Headers   []McpServerHeaderInput `json:"headers,omitempty" jsonschema:"custom headers for MCP server requests"`
}

type UpdateMcpServerInput struct {
	ID        string                  `json:"id" jsonschema:"MCP server ID (required)"`
	Name      *string                 `json:"name,omitempty" jsonschema:"new server name"`
	Transport *string                 `json:"transport,omitempty" jsonschema:"new transport type"`
	URL       *string                 `json:"url,omitempty" jsonschema:"new server URL"`
	Headers   *[]McpServerHeaderInput `json:"headers,omitempty" jsonschema:"new headers"`
}

type DeleteMcpServerInput struct {
	ID string `json:"id" jsonschema:"MCP server ID to delete"`
}

type ListMcpServerToolsInput struct {
	ID string `json:"id" jsonschema:"MCP server ID to list tools from"`
}

type McpServerToolOutput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ListMcpServerToolsOutput struct {
	Tools []McpServerToolOutput `json:"tools"`
}

// --- Tool registration ---

func (h *Handler) registerMcpServerTools(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "mcp_servers_list",
		Description: "List all configured MCP tool servers with pagination.",
	}, h.listMcpServers)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "mcp_servers_get",
		Description: "Get an MCP tool server by ID.",
	}, h.getMcpServer)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "mcp_servers_create",
		Description: "Register a new MCP tool server.",
	}, h.createMcpServer)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "mcp_servers_update",
		Description: "Update an existing MCP tool server. Only provided fields are changed.",
	}, h.updateMcpServer)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "mcp_servers_delete",
		Description: "Delete an MCP tool server by ID.",
	}, h.deleteMcpServer)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "mcp_servers_list_tools",
		Description: "Connect to an MCP tool server and list its available tools.",
	}, h.listMcpServerToolsHandler)
}

// --- Tool handlers ---

func (h *Handler) listMcpServers(ctx context.Context, _ *mcp.CallToolRequest, input ListMcpServersInput) (*mcp.CallToolResult, ListMcpServersOutput, error) {
	result, err := h.mcpServerApp.Queries.ListMcpServers.Execute(ctx, queries.ListMcpServersParams{
		Limit:  input.Limit,
		Offset: input.Offset,
	})
	if err != nil {
		return nil, ListMcpServersOutput{}, fmt.Errorf("list mcp servers failed: %w", err)
	}
	servers := make([]McpServerOutput, len(result.McpServers))
	for i, s := range result.McpServers {
		servers[i] = toMcpServerOutput(s)
	}
	return nil, ListMcpServersOutput{McpServers: servers, Total: result.Total}, nil
}

func (h *Handler) getMcpServer(ctx context.Context, _ *mcp.CallToolRequest, input GetMcpServerInput) (*mcp.CallToolResult, McpServerOutput, error) {
	result, err := h.mcpServerApp.Queries.GetMcpServer.Execute(ctx, queries.GetMcpServerParams{ID: input.ID})
	if err != nil {
		return nil, McpServerOutput{}, fmt.Errorf("get mcp server failed: %w", err)
	}
	return nil, toMcpServerOutput(result.McpServer), nil
}

func (h *Handler) createMcpServer(ctx context.Context, _ *mcp.CallToolRequest, input CreateMcpServerInput) (*mcp.CallToolResult, McpServerOutput, error) {
	headers := make([]mcpserver.McpServerHeader, len(input.Headers))
	for i, hdr := range input.Headers {
		headers[i] = mcpserver.McpServerHeader{Name: hdr.Name, Value: hdr.Value}
	}
	result, err := h.mcpServerApp.Commands.CreateMcpServer.Execute(ctx, commands.CreateMcpServerParams{
		Name:      input.Name,
		Transport: input.Transport,
		URL:       input.URL,
		Headers:   headers,
	})
	if err != nil {
		return nil, McpServerOutput{}, fmt.Errorf("create mcp server failed: %w", err)
	}
	return nil, toMcpServerOutput(result.McpServer), nil
}

func (h *Handler) updateMcpServer(ctx context.Context, _ *mcp.CallToolRequest, input UpdateMcpServerInput) (*mcp.CallToolResult, McpServerOutput, error) {
	params := commands.UpdateMcpServerParams{
		ID:        input.ID,
		Name:      input.Name,
		Transport: input.Transport,
		URL:       input.URL,
	}
	if input.Headers != nil {
		headers := make([]mcpserver.McpServerHeader, len(*input.Headers))
		for i, hdr := range *input.Headers {
			headers[i] = mcpserver.McpServerHeader{Name: hdr.Name, Value: hdr.Value}
		}
		params.Headers = &headers
	}
	result, err := h.mcpServerApp.Commands.UpdateMcpServer.Execute(ctx, params)
	if err != nil {
		return nil, McpServerOutput{}, fmt.Errorf("update mcp server failed: %w", err)
	}
	return nil, toMcpServerOutput(result.McpServer), nil
}

func (h *Handler) deleteMcpServer(ctx context.Context, _ *mcp.CallToolRequest, input DeleteMcpServerInput) (*mcp.CallToolResult, any, error) {
	err := h.mcpServerApp.Commands.DeleteMcpServer.Execute(ctx, commands.DeleteMcpServerParams{ID: input.ID})
	if err != nil {
		return nil, nil, fmt.Errorf("delete mcp server failed: %w", err)
	}
	return nil, map[string]string{"status": "deleted"}, nil
}

func (h *Handler) listMcpServerToolsHandler(ctx context.Context, _ *mcp.CallToolRequest, input ListMcpServerToolsInput) (*mcp.CallToolResult, ListMcpServerToolsOutput, error) {
	result, err := h.mcpServerApp.Queries.GetMcpServer.Execute(ctx, queries.GetMcpServerParams{ID: input.ID})
	if err != nil {
		return nil, ListMcpServerToolsOutput{}, fmt.Errorf("get mcp server failed: %w", err)
	}

	server := result.McpServer
	headerMap := make(map[string]string, len(server.Headers))
	for _, hdr := range server.Headers {
		headerMap[hdr.Name] = hdr.Value
	}

	tools, err := h.mcpDiscovery.ListTools(ctx, server.Name, server.Transport, server.URL, headerMap)
	if err != nil {
		// Return empty list on discovery failure (matches HTTP handler behavior)
		return nil, ListMcpServerToolsOutput{Tools: []McpServerToolOutput{}}, nil
	}

	out := make([]McpServerToolOutput, len(tools))
	for i, t := range tools {
		out[i] = McpServerToolOutput{Name: t.Name, Description: t.Description}
	}
	return nil, ListMcpServerToolsOutput{Tools: out}, nil
}

func toMcpServerOutput(s *mcpserver.McpServer) McpServerOutput {
	headers := make([]McpServerHeaderOutput, len(s.Headers))
	for i, hdr := range s.Headers {
		headers[i] = McpServerHeaderOutput{Name: hdr.Name, Value: hdr.Value}
	}
	return McpServerOutput{
		ID:        s.ID,
		Name:      s.Name,
		Transport: s.Transport,
		URL:       s.URL,
		Headers:   headers,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}
