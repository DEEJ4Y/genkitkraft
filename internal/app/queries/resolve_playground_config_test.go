package queries_test

import (
	"context"
	"testing"

	"github.com/DEEJ4Y/genkitkraft/internal/app/queries"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/agent"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/provider"
	"github.com/DEEJ4Y/genkitkraft/resources/test/mock"
)

// AgentID must reach the resulting ChatRequest regardless of caller — it's
// the only way the report_gap tool can be scoped to an agent when no
// session exists (the stateless deploy chat-completions path).
func TestResolvePlaygroundConfig_PopulatesAgentIDOnChatRequest(t *testing.T) {
	agentRepo := &mock.AgentRepository{
		GetByIDResult: &agent.Agent{ID: "agent-1", ProviderID: "provider-1", ModelID: "gpt-4o-mini", GapReportingEnabled: true},
	}
	providerRepo := &mock.ProviderRepository{
		GetByIDResult: &provider.Provider{ID: "provider-1", ProviderType: provider.OpenAI, BaseURL: "https://api.openai.com/v1"},
	}
	q := queries.NewResolvePlaygroundConfigQuery(agentRepo, providerRepo, &mock.PromptRepository{}, nil, nil, nil, nil)

	result, err := q.Execute(context.Background(), queries.ResolvePlaygroundConfigParams{AgentID: "agent-1"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if result.ChatRequest.AgentID != "agent-1" {
		t.Errorf("ChatRequest.AgentID = %q, want %q", result.ChatRequest.AgentID, "agent-1")
	}
	if !result.ChatRequest.GapReportingEnabled {
		t.Error("ChatRequest.GapReportingEnabled = false, want true (mirrors agent's flag)")
	}
	if result.ChatRequest.SessionID != "" {
		t.Errorf("ChatRequest.SessionID = %q, want empty — SessionID is set by the caller, not this query", result.ChatRequest.SessionID)
	}
}
