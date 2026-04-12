package agentrepo

import (
	"context"

	"github.com/DEEJ4Y/genkitkraft/internal/domain/agent"
)

// AgentRepository defines the contract for agent persistence.
type AgentRepository interface {
	List(ctx context.Context, limit, offset int) ([]*agent.Agent, error)
	Count(ctx context.Context) (int, error)
	GetByID(ctx context.Context, id string) (*agent.Agent, error)
	Create(ctx context.Context, a *agent.Agent) error
	Update(ctx context.Context, a *agent.Agent) error
	Delete(ctx context.Context, id string) error
}
