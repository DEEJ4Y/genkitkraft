package mcphandler

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/DEEJ4Y/genkitkraft/internal/app/commands"
	"github.com/DEEJ4Y/genkitkraft/internal/app/queries"
)

// --- Input/Output types ---

type LoginInput struct {
	Username string `json:"username" jsonschema:"username for authentication"`
	Password string `json:"password" jsonschema:"password for authentication"`
}

type LoginOutput struct {
	Token    string `json:"token"`
	Username string `json:"username"`
}

type LogoutInput struct {
	Token string `json:"token" jsonschema:"session token to invalidate"`
}

type GetMeInput struct {
	Token string `json:"token" jsonschema:"session token"`
}

type GetMeOutput struct {
	Username string `json:"username"`
}

type AuthStatusOutput struct {
	Required bool `json:"required"`
}

// --- Tool registration ---

func (h *Handler) registerAuthTools(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "auth_login",
		Description: "Authenticate with username and password. Returns a session token.",
	}, h.authLogin)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "auth_logout",
		Description: "Invalidate a session token.",
	}, h.authLogout)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "auth_get_me",
		Description: "Get the username associated with a session token.",
	}, h.authGetMe)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "auth_get_status",
		Description: "Check whether authentication is required.",
	}, h.authGetStatus)
}

// --- Tool handlers ---

func (h *Handler) authLogin(ctx context.Context, _ *mcp.CallToolRequest, input LoginInput) (*mcp.CallToolResult, LoginOutput, error) {
	result, err := h.authApp.Commands.Login.Execute(ctx, commands.LoginParams{
		Username: input.Username,
		Password: input.Password,
	})
	if err != nil {
		return nil, LoginOutput{}, fmt.Errorf("login failed: %w", err)
	}
	return nil, LoginOutput{Token: result.Token, Username: result.Username}, nil
}

func (h *Handler) authLogout(ctx context.Context, _ *mcp.CallToolRequest, input LogoutInput) (*mcp.CallToolResult, any, error) {
	err := h.authApp.Commands.Logout.Execute(ctx, commands.LogoutParams{Token: input.Token})
	if err != nil {
		return nil, nil, fmt.Errorf("logout failed: %w", err)
	}
	return nil, map[string]string{"status": "ok"}, nil
}

func (h *Handler) authGetMe(ctx context.Context, _ *mcp.CallToolRequest, input GetMeInput) (*mcp.CallToolResult, GetMeOutput, error) {
	result, err := h.authApp.Queries.GetMe.Execute(ctx, queries.GetMeParams{Token: input.Token})
	if err != nil {
		return nil, GetMeOutput{}, fmt.Errorf("get me failed: %w", err)
	}
	return nil, GetMeOutput{Username: result.Username}, nil
}

func (h *Handler) authGetStatus(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, AuthStatusOutput, error) {
	result, err := h.authApp.Queries.GetAuthStatus.Execute(ctx, queries.GetAuthStatusParams{})
	if err != nil {
		return nil, AuthStatusOutput{}, fmt.Errorf("get auth status failed: %w", err)
	}
	return nil, AuthStatusOutput{Required: result.Required}, nil
}
