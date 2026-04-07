package queries

import (
	"context"

	"github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	chatprovider "github.com/DEEJ4Y/genkitkraft/internal/ports/chat_provider"
	"github.com/DEEJ4Y/genkitkraft/internal/ports/encryptor"
	agentrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/agent_repo"
	promptrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/prompt_repo"
	providerrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/provider_repo"
)

type ResolvePlaygroundConfigParams struct {
	AgentID string
	// Optional overrides — empty/zero means use agent defaults
	ProviderID     string
	ModelID        string
	SystemPromptID *string // nil = use agent default, pointer to empty = clear prompt
	Temperature    *float64
	TopP           *float64
	TopK           *int
}

type ResolvePlaygroundConfigResult struct {
	ChatRequest chatprovider.ChatRequest
}

type ResolvePlaygroundConfigQuery struct {
	agentRepo    agentrepo.AgentRepository
	providerRepo providerrepo.ProviderRepository
	promptRepo   promptrepo.PromptRepository
	enc          encryptor.Encryptor
}

func NewResolvePlaygroundConfigQuery(
	agentRepo agentrepo.AgentRepository,
	providerRepo providerrepo.ProviderRepository,
	promptRepo promptrepo.PromptRepository,
	enc encryptor.Encryptor,
) *ResolvePlaygroundConfigQuery {
	return &ResolvePlaygroundConfigQuery{
		agentRepo:    agentRepo,
		providerRepo: providerRepo,
		promptRepo:   promptRepo,
		enc:          enc,
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
	topP := a.TopP
	if params.TopP != nil {
		topP = *params.TopP
	}
	topK := a.TopK
	if params.TopK != nil {
		topK = *params.TopK
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
		ProviderType: string(p.ProviderType),
		APIKey:       apiKey,
		BaseURL:      p.BaseURL,
		Config:       p.RawConfig,
		ModelID:      modelID,
		SystemPrompt: systemPrompt,
		Temperature:  temperature,
		TopP:         topP,
		TopK:         topK,
	}

	return ResolvePlaygroundConfigResult{ChatRequest: chatReq}, nil
}
