package queries_test

import (
	"context"
	"strings"
	"testing"

	"github.com/DEEJ4Y/genkitkraft/internal/app/queries"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/agent"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/prompt"
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

// Gap-reporting agents under-report in practice unless the system prompt
// itself carries usage instructions (see PR #49 manual test report) — the
// report_gap tool description alone isn't enough. So the instructions must
// reach every agent with the flag on, even one with no custom system prompt.
func TestResolvePlaygroundConfig_AppendsGapReportingInstructions_WhenEnabledAndNoCustomPrompt(t *testing.T) {
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
	if !strings.Contains(result.ChatRequest.SystemPrompt, "report_gap") {
		t.Errorf("ChatRequest.SystemPrompt = %q, want it to contain gap-reporting instructions", result.ChatRequest.SystemPrompt)
	}
}

// The instructions must be appended after, not instead of, an agent's own
// custom system prompt.
func TestResolvePlaygroundConfig_AppendsGapReportingInstructions_AfterCustomPrompt(t *testing.T) {
	const customPrompt = "You are a helpful assistant."
	agentRepo := &mock.AgentRepository{
		GetByIDResult: &agent.Agent{ID: "agent-1", ProviderID: "provider-1", ModelID: "gpt-4o-mini", SystemPromptID: "prompt-1", GapReportingEnabled: true},
	}
	providerRepo := &mock.ProviderRepository{
		GetByIDResult: &provider.Provider{ID: "provider-1", ProviderType: provider.OpenAI, BaseURL: "https://api.openai.com/v1"},
	}
	promptRepo := &mock.PromptRepository{GetByIDResult: &prompt.Prompt{ID: "prompt-1", Content: customPrompt}}
	q := queries.NewResolvePlaygroundConfigQuery(agentRepo, providerRepo, promptRepo, nil, nil, nil, nil)

	result, err := q.Execute(context.Background(), queries.ResolvePlaygroundConfigParams{AgentID: "agent-1"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	sp := result.ChatRequest.SystemPrompt
	if !strings.HasPrefix(sp, customPrompt) {
		t.Errorf("ChatRequest.SystemPrompt = %q, want it to start with the custom prompt %q", sp, customPrompt)
	}
	if idx := strings.Index(sp, "report_gap"); idx == -1 || idx < len(customPrompt) {
		t.Errorf("ChatRequest.SystemPrompt = %q, want gap-reporting instructions to follow the custom prompt", sp)
	}
}

// Agents without gap reporting enabled must not carry any of this
// instructional text — it would be meaningless without the report_gap tool.
func TestResolvePlaygroundConfig_NoGapReportingInstructions_WhenDisabled(t *testing.T) {
	const customPrompt = "You are a helpful assistant."
	agentRepo := &mock.AgentRepository{
		GetByIDResult: &agent.Agent{ID: "agent-1", ProviderID: "provider-1", ModelID: "gpt-4o-mini", SystemPromptID: "prompt-1", GapReportingEnabled: false},
	}
	providerRepo := &mock.ProviderRepository{
		GetByIDResult: &provider.Provider{ID: "provider-1", ProviderType: provider.OpenAI, BaseURL: "https://api.openai.com/v1"},
	}
	promptRepo := &mock.PromptRepository{GetByIDResult: &prompt.Prompt{ID: "prompt-1", Content: customPrompt}}
	q := queries.NewResolvePlaygroundConfigQuery(agentRepo, providerRepo, promptRepo, nil, nil, nil, nil)

	result, err := q.Execute(context.Background(), queries.ResolvePlaygroundConfigParams{AgentID: "agent-1"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if result.ChatRequest.SystemPrompt != customPrompt {
		t.Errorf("ChatRequest.SystemPrompt = %q, want exactly %q with nothing appended", result.ChatRequest.SystemPrompt, customPrompt)
	}
}
