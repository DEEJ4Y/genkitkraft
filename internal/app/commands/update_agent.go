package commands

import (
	"context"

	"github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/agent"
	agentrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/agent_repo"
	promptrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/prompt_repo"
	providerrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/provider_repo"
)

type UpdateAgentParams struct {
	ID             string
	Name           *string
	ProviderID     *string
	ModelID        *string
	SystemPromptID *string // nil=don't change, ""=clear, "uuid"=set
	Temperature    *float64
	TopP           *float64
	TopK           *int
}

type UpdateAgentResult struct {
	Agent *agent.Agent
}

type UpdateAgentCommand struct {
	repo         agentrepo.AgentRepository
	providerRepo providerrepo.ProviderRepository
	promptRepo   promptrepo.PromptRepository
}

func NewUpdateAgentCommand(repo agentrepo.AgentRepository, providerRepo providerrepo.ProviderRepository, promptRepo promptrepo.PromptRepository) *UpdateAgentCommand {
	return &UpdateAgentCommand{repo: repo, providerRepo: providerRepo, promptRepo: promptRepo}
}

func (c *UpdateAgentCommand) Execute(ctx context.Context, params UpdateAgentParams) (UpdateAgentResult, error) {
	a, err := c.repo.GetByID(ctx, params.ID)
	if err != nil {
		return UpdateAgentResult{}, err
	}

	if params.Name != nil {
		a.Name = *params.Name
	}
	if params.ProviderID != nil {
		// Validate provider exists
		if _, err := c.providerRepo.GetByID(ctx, *params.ProviderID); err != nil {
			return UpdateAgentResult{}, errors.NewAppError(errors.InvalidInput, "provider not found")
		}
		a.ProviderID = *params.ProviderID
	}
	if params.ModelID != nil {
		a.ModelID = *params.ModelID
	}
	if params.SystemPromptID != nil {
		if *params.SystemPromptID != "" {
			// Validate prompt exists
			if _, err := c.promptRepo.GetByID(ctx, *params.SystemPromptID); err != nil {
				return UpdateAgentResult{}, errors.NewAppError(errors.InvalidInput, "system prompt not found")
			}
		}
		a.SystemPromptID = *params.SystemPromptID
	}
	if params.Temperature != nil {
		a.Temperature = *params.Temperature
	}
	if params.TopP != nil {
		a.TopP = *params.TopP
	}
	if params.TopK != nil {
		a.TopK = *params.TopK
	}

	if err := c.repo.Update(ctx, a); err != nil {
		return UpdateAgentResult{}, err
	}

	// Re-fetch to populate resolved fields
	a, err = c.repo.GetByID(ctx, a.ID)
	if err != nil {
		return UpdateAgentResult{}, err
	}

	return UpdateAgentResult{Agent: a}, nil
}
