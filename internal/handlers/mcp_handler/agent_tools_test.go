package mcphandler

import (
	"context"
	"testing"

	"github.com/DEEJ4Y/genkitkraft/internal/app"
	"github.com/DEEJ4Y/genkitkraft/internal/app/commands"
	"github.com/DEEJ4Y/genkitkraft/internal/app/queries"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/agent"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/provider"
	"github.com/DEEJ4Y/genkitkraft/resources/test/mock"
)

func newAgentTestHandler(repo *mock.AgentRepository) *Handler {
	providerRepo := &mock.ProviderRepository{GetByIDResult: &provider.Provider{ID: "provider-1"}}
	promptRepo := &mock.PromptRepository{}

	createAgentCmd := commands.NewCreateAgentCommand(repo, providerRepo, promptRepo)
	updateAgentCmd := commands.NewUpdateAgentCommand(repo, providerRepo, promptRepo)

	return &Handler{
		agentApp: &app.AgentApp{
			Commands: app.AgentCommands{
				CreateAgent: createAgentCmd,
				UpdateAgent: updateAgentCmd,
			},
			Queries: app.AgentQueries{
				ListAgents: queries.NewListAgentsQuery(repo),
				GetAgent:   queries.NewGetAgentQuery(repo),
			},
		},
	}
}

func TestCreateAgentTool_RoundTripsMaxToolCallsAndGapReporting(t *testing.T) {
	repo := &mock.AgentRepository{}
	h := newAgentTestHandler(repo)

	maxToolCalls := 7
	gapReportingEnabled := true

	_, out, err := h.createAgent(context.Background(), nil, CreateAgentInput{
		Name:                "test-agent",
		ProviderID:          "provider-1",
		ModelID:             "gpt-4o",
		MaxToolCalls:        &maxToolCalls,
		GapReportingEnabled: &gapReportingEnabled,
	})
	if err != nil {
		t.Fatalf("createAgent: %v", err)
	}
	if out.MaxToolCalls != 7 {
		t.Errorf("out.MaxToolCalls = %d, want 7", out.MaxToolCalls)
	}
	if !out.GapReportingEnabled {
		t.Errorf("out.GapReportingEnabled = false, want true")
	}
	if repo.LastCreate == nil || repo.LastCreate.MaxToolCalls != 7 || !repo.LastCreate.GapReportingEnabled {
		t.Errorf("repo.LastCreate = %+v, want MaxToolCalls=7 GapReportingEnabled=true", repo.LastCreate)
	}
}

func TestUpdateAgentTool_RoundTripsMaxToolCallsAndGapReporting(t *testing.T) {
	repo := &mock.AgentRepository{GetByIDResult: &agent.Agent{ID: "agent-1", MaxToolCalls: 10, GapReportingEnabled: false}}
	h := newAgentTestHandler(repo)

	maxToolCalls := 3
	gapReportingEnabled := true

	_, out, err := h.updateAgent(context.Background(), nil, UpdateAgentInput{
		ID:                  "agent-1",
		MaxToolCalls:        &maxToolCalls,
		GapReportingEnabled: &gapReportingEnabled,
	})
	if err != nil {
		t.Fatalf("updateAgent: %v", err)
	}
	if out.MaxToolCalls != 3 {
		t.Errorf("out.MaxToolCalls = %d, want 3", out.MaxToolCalls)
	}
	if !out.GapReportingEnabled {
		t.Errorf("out.GapReportingEnabled = false, want true")
	}
	if repo.LastUpdate == nil || repo.LastUpdate.MaxToolCalls != 3 || !repo.LastUpdate.GapReportingEnabled {
		t.Errorf("repo.LastUpdate = %+v, want MaxToolCalls=3 GapReportingEnabled=true", repo.LastUpdate)
	}
}

func TestGetAgentTool_SurfacesMaxToolCallsAndGapReporting(t *testing.T) {
	repo := &mock.AgentRepository{GetByIDResult: &agent.Agent{ID: "agent-1", MaxToolCalls: 5, GapReportingEnabled: true}}
	h := newAgentTestHandler(repo)

	_, out, err := h.getAgent(context.Background(), nil, GetAgentInput{ID: "agent-1"})
	if err != nil {
		t.Fatalf("getAgent: %v", err)
	}
	if out.MaxToolCalls != 5 || !out.GapReportingEnabled {
		t.Errorf("out = %+v, want MaxToolCalls=5 GapReportingEnabled=true", out)
	}
}

func TestListAgentsTool_SurfacesMaxToolCallsAndGapReporting(t *testing.T) {
	repo := &mock.AgentRepository{
		ListAgents:  []*agent.Agent{{ID: "agent-1", MaxToolCalls: 8, GapReportingEnabled: true}},
		CountResult: 1,
	}
	h := newAgentTestHandler(repo)

	_, out, err := h.listAgents(context.Background(), nil, ListAgentsInput{})
	if err != nil {
		t.Fatalf("listAgents: %v", err)
	}
	if len(out.Agents) != 1 || out.Agents[0].MaxToolCalls != 8 || !out.Agents[0].GapReportingEnabled {
		t.Errorf("out.Agents = %+v, want a single agent with MaxToolCalls=8 GapReportingEnabled=true", out.Agents)
	}
}
