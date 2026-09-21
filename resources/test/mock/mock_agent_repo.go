package mock

import (
	"context"

	"github.com/DEEJ4Y/genkitkraft/internal/domain/agent"
	agentrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/agent_repo"
)

// Compile-time check.
var _ agentrepo.AgentRepository = (*AgentRepository)(nil)

// AgentRepository is a mock implementation of the AgentRepository port for testing.
type AgentRepository struct {
	// ListAgents is returned by List(). Set before calling.
	ListAgents []*agent.Agent
	ListErr    error

	CountResult int
	CountErr    error

	GetByIDResult *agent.Agent
	GetByIDErr    error

	CreateErr error
	UpdateErr error
	DeleteErr error

	// LastCreate/LastUpdate capture the most recent call's argument for assertions.
	LastCreate *agent.Agent
	LastUpdate *agent.Agent

	LastListLimit  int
	LastListOffset int
	LastGetByID    string
	LastDelete     string
}

func (m *AgentRepository) List(_ context.Context, limit, offset int) ([]*agent.Agent, error) {
	m.LastListLimit = limit
	m.LastListOffset = offset
	return m.ListAgents, m.ListErr
}

func (m *AgentRepository) Count(_ context.Context) (int, error) {
	return m.CountResult, m.CountErr
}

func (m *AgentRepository) GetByID(_ context.Context, id string) (*agent.Agent, error) {
	m.LastGetByID = id
	return m.GetByIDResult, m.GetByIDErr
}

// Create, like the real adapters, makes the agent immediately visible to a
// subsequent GetByID — CreateAgentCommand/UpdateAgentCommand re-fetch after
// writing to populate resolved (joined) fields.
func (m *AgentRepository) Create(_ context.Context, a *agent.Agent) error {
	if a.ID == "" {
		a.ID = "mock-agent-id"
	}
	m.LastCreate = a
	if m.CreateErr == nil {
		m.GetByIDResult = a
	}
	return m.CreateErr
}

func (m *AgentRepository) Update(_ context.Context, a *agent.Agent) error {
	m.LastUpdate = a
	if m.UpdateErr == nil {
		m.GetByIDResult = a
	}
	return m.UpdateErr
}

func (m *AgentRepository) Delete(_ context.Context, id string) error {
	m.LastDelete = id
	return m.DeleteErr
}
