package commands

import (
	"context"

	"github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/agent"
	agentrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/agent_repo"
	promptrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/prompt_repo"
	providerrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/provider_repo"
)

type CreateAgentParams struct {
	Name           string
	ProviderID     string
	ModelID        string
	SystemPromptID string
	Temperature    *float64
	TopP           *float64
	TopK           *int
}

type CreateAgentResult struct {
	Agent *agent.Agent
}

type CreateAgentCommand struct {
	repo         agentrepo.AgentRepository
	providerRepo providerrepo.ProviderRepository
	promptRepo   promptrepo.PromptRepository
}

func NewCreateAgentCommand(repo agentrepo.AgentRepository, providerRepo providerrepo.ProviderRepository, promptRepo promptrepo.PromptRepository) *CreateAgentCommand {
	return &CreateAgentCommand{repo: repo, providerRepo: providerRepo, promptRepo: promptRepo}
}

func (c *CreateAgentCommand) Execute(ctx context.Context, params CreateAgentParams) (CreateAgentResult, error) {
	if params.Name == "" {
		return CreateAgentResult{}, errors.NewAppError(errors.InvalidInput, "name is required")
	}
	if params.ProviderID == "" {
		return CreateAgentResult{}, errors.NewAppError(errors.InvalidInput, "provider is required")
	}
	if params.ModelID == "" {
		return CreateAgentResult{}, errors.NewAppError(errors.InvalidInput, "model is required")
	}

	// Validate provider exists
	if _, err := c.providerRepo.GetByID(ctx, params.ProviderID); err != nil {
		return CreateAgentResult{}, errors.NewAppError(errors.InvalidInput, "provider not found")
	}

	// Validate system prompt exists if provided
	if params.SystemPromptID != "" {
		if _, err := c.promptRepo.GetByID(ctx, params.SystemPromptID); err != nil {
			return CreateAgentResult{}, errors.NewAppError(errors.InvalidInput, "system prompt not found")
		}
	}

	temperature := agent.DefaultTemperature
	if params.Temperature != nil {
		temperature = *params.Temperature
	}
	topP := agent.DefaultTopP
	if params.TopP != nil {
		topP = *params.TopP
	}
	topK := agent.DefaultTopK
	if params.TopK != nil {
		topK = *params.TopK
	}

	a := &agent.Agent{
		Name:           params.Name,
		ProviderID:     params.ProviderID,
		ModelID:        params.ModelID,
		SystemPromptID: params.SystemPromptID,
		Temperature:    temperature,
		TopP:           topP,
		TopK:           topK,
	}

	if err := c.repo.Create(ctx, a); err != nil {
		return CreateAgentResult{}, err
	}

	// Re-fetch to populate resolved fields
	a, err := c.repo.GetByID(ctx, a.ID)
	if err != nil {
		return CreateAgentResult{}, err
	}

	return CreateAgentResult{Agent: a}, nil
}
