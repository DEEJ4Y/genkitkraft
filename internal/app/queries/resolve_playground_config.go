package queries

import (
	"context"
	"encoding/json"

	"github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	chatprovider "github.com/DEEJ4Y/genkitkraft/internal/ports/chat_provider"
	"github.com/DEEJ4Y/genkitkraft/internal/ports/encryptor"
	agentrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/agent_repo"
	agenttoolrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/agent_tool_repo"
	httptoolrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/http_tool_repo"
	mcpserverrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/mcp_server_repo"
	promptrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/prompt_repo"
	providerrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/provider_repo"
)

// ToolOverride specifies tool configuration overrides for a chat request.
type ToolOverride struct {
	HttpToolIDs    []string
	McpServers     []agenttoolrepo.McpServerToolConfig
	BuiltInToolIDs []string
}

type ResolvePlaygroundConfigParams struct {
	AgentID string
	// Optional overrides — empty/zero means use agent defaults
	ProviderID         string
	ModelID            string
	SystemPromptID     *string // nil = use agent default, pointer to empty = clear prompt
	TemperatureEnabled *bool
	Temperature        *float64
	TopPEnabled        *bool
	TopP               *float64
	TopKEnabled        *bool
	TopK               *int
	MaxToolCalls       *int
	// Optional tool overrides — nil means use agent defaults
	ToolOverride *ToolOverride
	// IncludeTools controls whether tools are resolved. Deploy and playground set this true.
	IncludeTools bool
}

type ResolvePlaygroundConfigResult struct {
	ChatRequest chatprovider.ChatRequest
}

type ResolvePlaygroundConfigQuery struct {
	agentRepo     agentrepo.AgentRepository
	providerRepo  providerrepo.ProviderRepository
	promptRepo    promptrepo.PromptRepository
	enc           encryptor.Encryptor
	agentToolRepo agenttoolrepo.AgentToolRepository
	httpToolRepo  httptoolrepo.HttpToolRepository
	mcpServerRepo mcpserverrepo.McpServerRepository
}

func NewResolvePlaygroundConfigQuery(
	agentRepo agentrepo.AgentRepository,
	providerRepo providerrepo.ProviderRepository,
	promptRepo promptrepo.PromptRepository,
	enc encryptor.Encryptor,
	agentToolRepo agenttoolrepo.AgentToolRepository,
	httpToolRepo httptoolrepo.HttpToolRepository,
	mcpServerRepo mcpserverrepo.McpServerRepository,
) *ResolvePlaygroundConfigQuery {
	return &ResolvePlaygroundConfigQuery{
		agentRepo:     agentRepo,
		providerRepo:  providerRepo,
		promptRepo:    promptRepo,
		enc:           enc,
		agentToolRepo: agentToolRepo,
		httpToolRepo:  httpToolRepo,
		mcpServerRepo: mcpServerRepo,
	}
}

func (q *ResolvePlaygroundConfigQuery) Execute(ctx context.Context, params ResolvePlaygroundConfigParams) (ResolvePlaygroundConfigResult, error) {
	// Load agent
	a, err := q.agentRepo.GetByID(ctx, params.AgentID)
	if err != nil {
		return ResolvePlaygroundConfigResult{}, err
	}

	// Determine effective config (apply overrides)
	providerID := a.ProviderID
	if params.ProviderID != "" {
		providerID = params.ProviderID
	}
	modelID := a.ModelID
	if params.ModelID != "" {
		modelID = params.ModelID
	}
	temperature := a.Temperature
	if params.Temperature != nil {
		temperature = *params.Temperature
	}
	temperatureEnabled := a.TemperatureEnabled
	if params.TemperatureEnabled != nil {
		temperatureEnabled = *params.TemperatureEnabled
	}
	topP := a.TopP
	if params.TopP != nil {
		topP = *params.TopP
	}
	topPEnabled := a.TopPEnabled
	if params.TopPEnabled != nil {
		topPEnabled = *params.TopPEnabled
	}
	topK := a.TopK
	if params.TopK != nil {
		topK = *params.TopK
	}
	topKEnabled := a.TopKEnabled
	if params.TopKEnabled != nil {
		topKEnabled = *params.TopKEnabled
	}
	maxToolCalls := a.MaxToolCalls
	if params.MaxToolCalls != nil {
		maxToolCalls = *params.MaxToolCalls
	}

	// Determine system prompt ID
	systemPromptID := a.SystemPromptID
	if params.SystemPromptID != nil {
		systemPromptID = *params.SystemPromptID
	}

	// Load provider
	p, err := q.providerRepo.GetByID(ctx, providerID)
	if err != nil {
		return ResolvePlaygroundConfigResult{}, errors.NewAppError(errors.InvalidInput, "provider not found")
	}

	// Decrypt API key
	apiKey := ""
	if p.APIKey != nil {
		decrypted, err := q.enc.Decrypt(*p.APIKey)
		if err != nil {
			return ResolvePlaygroundConfigResult{}, errors.NewAppErrorf(errors.Internal, "decrypting api key: %v", err)
		}
		apiKey = decrypted
	}

	// Load system prompt content
	systemPrompt := ""
	if systemPromptID != "" {
		prompt, err := q.promptRepo.GetByID(ctx, systemPromptID)
		if err != nil {
			return ResolvePlaygroundConfigResult{}, errors.NewAppError(errors.InvalidInput, "system prompt not found")
		}
		systemPrompt = prompt.Content
	}

	chatReq := chatprovider.ChatRequest{
		ProviderType:       string(p.ProviderType),
		APIKey:             apiKey,
		BaseURL:            p.BaseURL,
		Config:             p.RawConfig,
		ModelID:            modelID,
		SystemPrompt:       systemPrompt,
		TemperatureEnabled: temperatureEnabled,
		Temperature:        temperature,
		TopPEnabled:        topPEnabled,
		TopP:               topP,
		TopKEnabled:        topKEnabled,
		TopK:               topK,
		MaxToolCalls:       maxToolCalls,
	}

	// Resolve tools if requested
	if params.IncludeTools {
		if err := q.resolveTools(ctx, params, &chatReq); err != nil {
			return ResolvePlaygroundConfigResult{}, err
		}
	}

	return ResolvePlaygroundConfigResult{ChatRequest: chatReq}, nil
}

// resolveTools loads tool configurations and populates the ChatRequest.
func (q *ResolvePlaygroundConfigQuery) resolveTools(ctx context.Context, params ResolvePlaygroundConfigParams, chatReq *chatprovider.ChatRequest) error {
	var httpToolIDs []string
	var mcpConfigs []agenttoolrepo.McpServerToolConfig

	if params.ToolOverride != nil {
		// Use override values
		httpToolIDs = params.ToolOverride.HttpToolIDs
		mcpConfigs = params.ToolOverride.McpServers
		chatReq.BuiltInToolIDs = params.ToolOverride.BuiltInToolIDs
	} else {
		// Load agent's saved tool config
		agentTools, err := q.agentToolRepo.GetByAgentID(ctx, params.AgentID)
		if err != nil {
			return errors.NewAppErrorf(errors.Internal, "loading agent tools: %v", err)
		}
		httpToolIDs = agentTools.HttpToolIDs
		mcpConfigs = agentTools.McpServers
		chatReq.BuiltInToolIDs = agentTools.BuiltInToolIDs
	}

	// Resolve HTTP tools
	for _, toolID := range httpToolIDs {
		tool, err := q.httpToolRepo.GetByID(ctx, toolID)
		if err != nil {
			continue // Skip tools that no longer exist
		}
		headers := make(map[string]string, len(tool.Headers))
		for _, h := range tool.Headers {
			headers[h.Name] = h.Value
		}
		chatReq.HttpTools = append(chatReq.HttpTools, chatprovider.HttpToolDefinition{
			ID:           tool.ID,
			Name:         tool.Name,
			Description:  tool.Description,
			Method:       tool.Method,
			URL:          tool.URL,
			Headers:      headers,
			BodyTemplate: tool.BodyTemplate,
			InputSchema:  json.RawMessage(tool.InputSchema),
		})
	}

	// Resolve MCP servers
	for _, mc := range mcpConfigs {
		server, err := q.mcpServerRepo.GetByID(ctx, mc.McpServerID)
		if err != nil {
			continue // Skip servers that no longer exist
		}
		headers := make(map[string]string, len(server.Headers))
		for _, h := range server.Headers {
			headers[h.Name] = h.Value
		}
		chatReq.McpServers = append(chatReq.McpServers, chatprovider.McpServerConfig{
			ServerID:  server.ID,
			Transport: server.Transport,
			URL:       server.URL,
			Headers:   headers,
			SelectAll: mc.SelectAll,
			ToolNames: mc.ToolNames,
		})
	}

	return nil
}
