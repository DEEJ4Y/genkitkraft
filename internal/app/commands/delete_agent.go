package commands

import (
	"context"

	agentrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/agent_repo"
)

type DeleteAgentParams struct {
	ID string
}

type DeleteAgentCommand struct {
	repo agentrepo.AgentRepository
}

func NewDeleteAgentCommand(repo agentrepo.AgentRepository) *DeleteAgentCommand {
	return &DeleteAgentCommand{repo: repo}
}

func (c *DeleteAgentCommand) Execute(ctx context.Context, params DeleteAgentParams) error {
	return c.repo.Delete(ctx, params.ID)
}
