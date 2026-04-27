package mcphandler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/DEEJ4Y/genkitkraft/internal/app/commands"
	"github.com/DEEJ4Y/genkitkraft/internal/app/queries"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/provider"
)

// --- Input/Output types ---

type ListProvidersInput struct{}

type ProviderOutput struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	ProviderType string            `json:"provider_type"`
	APIKey       string            `json:"api_key,omitempty"`
	BaseURL      string            `json:"base_url,omitempty"`
	Config       map[string]string `json:"config,omitempty"`
	Enabled      bool              `json:"enabled"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

type ListProvidersOutput struct {
	Providers []ProviderOutput `json:"providers"`
}

type GetProviderInput struct {
	ID string `json:"id" jsonschema:"provider ID"`
}

type CreateProviderInput struct {
	Name         string            `json:"name" jsonschema:"provider name (required)"`
	ProviderType string            `json:"provider_type" jsonschema:"provider type, e.g. openai, google_ai, anthropic (required)"`
	APIKey       *string           `json:"api_key,omitempty" jsonschema:"API key for the provider"`
	BaseURL      string            `json:"base_url,omitempty" jsonschema:"custom base URL"`
	Config       map[string]string `json:"config,omitempty" jsonschema:"provider-specific config fields"`
}

type UpdateProviderInput struct {
	ID      string             `json:"id" jsonschema:"provider ID (required)"`
	Name    *string            `json:"name,omitempty" jsonschema:"new provider name"`
	APIKey  *string            `json:"api_key,omitempty" jsonschema:"new API key"`
	BaseURL *string            `json:"base_url,omitempty" jsonschema:"new base URL"`
	Config  *map[string]string `json:"config,omitempty" jsonschema:"new config fields"`
	Enabled *bool              `json:"enabled,omitempty" jsonschema:"enable or disable the provider"`
}

type DeleteProviderInput struct {
	ID string `json:"id" jsonschema:"provider ID to delete"`
}

type TestProviderInput struct {
	ID string `json:"id" jsonschema:"provider ID to test connectivity"`
}

type TestProviderOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ListProviderTypesInput struct{}

type ProviderTypeOutput struct {
	Type            string `json:"type"`
	DisplayName     string `json:"display_name"`
	RequiresAPIKey  bool   `json:"requires_api_key"`
	RequiresBaseURL bool   `json:"requires_base_url"`
	ModelPrefix     string `json:"model_prefix,omitempty"`
	BaseURLDefault  string `json:"base_url_default,omitempty"`
	ComingSoon      bool   `json:"coming_soon,omitempty"`
}

type ListProviderTypesOutput struct {
	ProviderTypes []ProviderTypeOutput `json:"provider_types"`
}

// --- Tool registration ---

func (h *Handler) registerProviderTools(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "providers_list",
		Description: "List all configured LLM providers.",
	}, h.listProviders)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "providers_get",
		Description: "Get an LLM provider by ID.",
	}, h.getProvider)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "providers_create",
		Description: "Create a new LLM provider configuration.",
	}, h.createProvider)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "providers_update",
		Description: "Update an existing LLM provider. Only provided fields are changed.",
	}, h.updateProvider)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "providers_delete",
		Description: "Delete an LLM provider by ID.",
	}, h.deleteProvider)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "providers_test",
		Description: "Test connectivity to an LLM provider.",
	}, h.testProvider)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "provider_types_list",
		Description: "List all supported LLM provider types and their metadata.",
	}, h.listProviderTypes)
}

// --- Tool handlers ---

func (h *Handler) listProviders(ctx context.Context, _ *mcp.CallToolRequest, _ ListProvidersInput) (*mcp.CallToolResult, ListProvidersOutput, error) {
	result, err := h.providerApp.Queries.ListProviders.Execute(ctx, queries.ListProvidersParams{})
	if err != nil {
		return nil, ListProvidersOutput{}, fmt.Errorf("list providers failed: %w", err)
	}
	providers := make([]ProviderOutput, len(result.Providers))
	for i, p := range result.Providers {
		providers[i] = toProviderOutput(p)
	}
	return nil, ListProvidersOutput{Providers: providers}, nil
}

func (h *Handler) getProvider(ctx context.Context, _ *mcp.CallToolRequest, input GetProviderInput) (*mcp.CallToolResult, ProviderOutput, error) {
	result, err := h.providerApp.Queries.GetProvider.Execute(ctx, queries.GetProviderParams{ID: input.ID})
	if err != nil {
		return nil, ProviderOutput{}, fmt.Errorf("get provider failed: %w", err)
	}
	return nil, toProviderOutput(result.Provider), nil
}

func (h *Handler) createProvider(ctx context.Context, _ *mcp.CallToolRequest, input CreateProviderInput) (*mcp.CallToolResult, ProviderOutput, error) {
	result, err := h.providerApp.Commands.CreateProvider.Execute(ctx, commands.CreateProviderParams{
		Name:         input.Name,
		ProviderType: provider.ProviderType(input.ProviderType),
		APIKey:       input.APIKey,
		BaseURL:      input.BaseURL,
		Config:       input.Config,
	})
	if err != nil {
		return nil, ProviderOutput{}, fmt.Errorf("create provider failed: %w", err)
	}
	return nil, toProviderOutput(result.Provider), nil
}

func (h *Handler) updateProvider(ctx context.Context, _ *mcp.CallToolRequest, input UpdateProviderInput) (*mcp.CallToolResult, ProviderOutput, error) {
	result, err := h.providerApp.Commands.UpdateProvider.Execute(ctx, commands.UpdateProviderParams{
		ID:      input.ID,
		Name:    input.Name,
		APIKey:  input.APIKey,
		BaseURL: input.BaseURL,
		Config:  input.Config,
		Enabled: input.Enabled,
	})
	if err != nil {
		return nil, ProviderOutput{}, fmt.Errorf("update provider failed: %w", err)
	}
	return nil, toProviderOutput(result.Provider), nil
}

func (h *Handler) deleteProvider(ctx context.Context, _ *mcp.CallToolRequest, input DeleteProviderInput) (*mcp.CallToolResult, any, error) {
	err := h.providerApp.Commands.DeleteProvider.Execute(ctx, commands.DeleteProviderParams{ID: input.ID})
	if err != nil {
		return nil, nil, fmt.Errorf("delete provider failed: %w", err)
	}
	return nil, map[string]string{"status": "deleted"}, nil
}

func (h *Handler) testProvider(ctx context.Context, _ *mcp.CallToolRequest, input TestProviderInput) (*mcp.CallToolResult, TestProviderOutput, error) {
	result, err := h.providerApp.Commands.TestProvider.Execute(ctx, commands.TestProviderParams{ID: input.ID})
	if err != nil {
		return nil, TestProviderOutput{}, fmt.Errorf("test provider failed: %w", err)
	}
	return nil, TestProviderOutput{Success: result.Success, Message: result.Message}, nil
}

func (h *Handler) listProviderTypes(ctx context.Context, _ *mcp.CallToolRequest, _ ListProviderTypesInput) (*mcp.CallToolResult, ListProviderTypesOutput, error) {
	result, err := h.providerApp.Queries.ListProviderTypes.Execute(ctx, queries.ListProviderTypesParams{})
	if err != nil {
		return nil, ListProviderTypesOutput{}, fmt.Errorf("list provider types failed: %w", err)
	}
	types := make([]ProviderTypeOutput, len(result.ProviderTypes))
	for i, pt := range result.ProviderTypes {
		types[i] = ProviderTypeOutput{
			Type:            string(pt.Type),
			DisplayName:     pt.DisplayName,
			RequiresAPIKey:  pt.RequiresAPIKey,
			RequiresBaseURL: pt.RequiresBaseURL,
			ModelPrefix:     pt.ModelPrefix,
			BaseURLDefault:  pt.BaseURLDefault,
			ComingSoon:      pt.ComingSoon,
		}
	}
	return nil, ListProviderTypesOutput{ProviderTypes: types}, nil
}

func toProviderOutput(p *provider.Provider) ProviderOutput {
	out := ProviderOutput{
		ID:           p.ID,
		Name:         p.Name,
		ProviderType: string(p.ProviderType),
		BaseURL:      p.BaseURL,
		Enabled:      p.Enabled,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}
	if masked := p.MaskedAPIKey(); masked != nil {
		out.APIKey = *masked
	}
	if len(p.RawConfig) > 0 && string(p.RawConfig) != "{}" {
		var configMap map[string]string
		if err := json.Unmarshal(p.RawConfig, &configMap); err == nil && len(configMap) > 0 {
			out.Config = configMap
		}
	}
	return out
}
