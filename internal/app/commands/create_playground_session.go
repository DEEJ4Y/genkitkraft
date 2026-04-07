package commands

import (
	"context"

	"github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/playground"
	agentrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/agent_repo"
	playgroundrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/playground_repo"
)

type CreatePlaygroundSessionParams struct {
	AgentID string
	Title   string
}

type CreatePlaygroundSessionResult struct {
	Session *playground.Session
}

type CreatePlaygroundSessionCommand struct {
	repo      playgroundrepo.PlaygroundRepository
	agentRepo agentrepo.AgentRepository
}

func NewCreatePlaygroundSessionCommand(repo playgroundrepo.PlaygroundRepository, agentRepo agentrepo.AgentRepository) *CreatePlaygroundSessionCommand {
	return &CreatePlaygroundSessionCommand{repo: repo, agentRepo: agentRepo}
}

func (c *CreatePlaygroundSessionCommand) Execute(ctx context.Context, params CreatePlaygroundSessionParams) (CreatePlaygroundSessionResult, error) {
	if params.AgentID == "" {
		return CreatePlaygroundSessionResult{}, errors.NewAppError(errors.InvalidInput, "agent ID is required")
	}

	// Validate agent exists
	if _, err := c.agentRepo.GetByID(ctx, params.AgentID); err != nil {
		return CreatePlaygroundSessionResult{}, err
	}

	s := &playground.Session{
		AgentID: params.AgentID,
		Title:   params.Title,
	}

	if err := c.repo.CreateSession(ctx, s); err != nil {
		return CreatePlaygroundSessionResult{}, err
	}

	return CreatePlaygroundSessionResult{Session: s}, nil
}
