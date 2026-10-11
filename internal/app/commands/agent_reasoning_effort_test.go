package commands

import (
	"context"
	"strings"
	"testing"

	apperrors "github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/agent"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/provider"
	"github.com/DEEJ4Y/genkitkraft/resources/test/mock"
)

func newReasoningCreateCommand() (*CreateAgentCommand, *mock.AgentRepository) {
	agentRepo := &mock.AgentRepository{}
	providerRepo := &mock.ProviderRepository{
		GetByIDResult: &provider.Provider{ID: "provider-1", ProviderType: provider.OpenAI},
	}
	return NewCreateAgentCommand(agentRepo, providerRepo, &mock.PromptRepository{}), agentRepo
}

func newReasoningUpdateCommand(a *agent.Agent) (*UpdateAgentCommand, *mock.AgentRepository) {
	agentRepo := &mock.AgentRepository{GetByIDResult: a}
	providerRepo := &mock.ProviderRepository{
		GetByIDResult: &provider.Provider{ID: "provider-1", ProviderType: provider.OpenAI},
	}
	return NewUpdateAgentCommand(agentRepo, providerRepo, &mock.PromptRepository{}), agentRepo
}

func reasoningCreateParams() CreateAgentParams {
	return CreateAgentParams{Name: "a", ProviderID: "provider-1", ModelID: "o3"}
}

func ptrTo[T any](v T) *T { return &v }

func assertInvalidInput(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected an error")
	}
	appErr, ok := apperrors.IsAppError(err)
	if !ok || appErr.Code() != apperrors.InvalidInput {
		t.Errorf("error = %v, want an InvalidInput app error", err)
	}
}

func TestCreateAgent_ReasoningEffort_DefaultsToDisabled(t *testing.T) {
	c, repo := newReasoningCreateCommand()

	if _, err := c.Execute(context.Background(), reasoningCreateParams()); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if repo.LastCreate.ReasoningEffortEnabled || repo.LastCreate.ReasoningEffort != "" {
		t.Errorf("reasoning = (%v, %q), want (false, \"\")", repo.LastCreate.ReasoningEffortEnabled, repo.LastCreate.ReasoningEffort)
	}
}

func TestCreateAgent_ReasoningEffort_StoresTrimmedValue(t *testing.T) {
	c, repo := newReasoningCreateCommand()
	p := reasoningCreateParams()
	p.ReasoningEffortEnabled = ptrTo(true)
	p.ReasoningEffort = ptrTo(" medium ")

	if _, err := c.Execute(context.Background(), p); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !repo.LastCreate.ReasoningEffortEnabled || repo.LastCreate.ReasoningEffort != "medium" {
		t.Errorf("reasoning = (%v, %q), want (true, %q)", repo.LastCreate.ReasoningEffortEnabled, repo.LastCreate.ReasoningEffort, "medium")
	}
}

func TestCreateAgent_ReasoningEffort_EnabledWithoutValueIsInvalid(t *testing.T) {
	c, repo := newReasoningCreateCommand()
	p := reasoningCreateParams()
	p.ReasoningEffortEnabled = ptrTo(true)
	p.ReasoningEffort = ptrTo("   ")

	_, err := c.Execute(context.Background(), p)
	assertInvalidInput(t, err)
	if repo.LastCreate != nil {
		t.Error("agent must not be created when the input is invalid")
	}
}

func TestCreateAgent_ReasoningEffort_TooLongIsInvalid(t *testing.T) {
	c, _ := newReasoningCreateCommand()
	p := reasoningCreateParams()
	p.ReasoningEffort = ptrTo(strings.Repeat("x", agent.MaxReasoningEffortLength+1))

	_, err := c.Execute(context.Background(), p)
	assertInvalidInput(t, err)
}

func TestUpdateAgent_ReasoningEffort_OmittedFieldsAreUnchanged(t *testing.T) {
	c, repo := newReasoningUpdateCommand(&agent.Agent{
		ID: "agent-1", ProviderID: "provider-1", ModelID: "o3",
		ReasoningEffortEnabled: true, ReasoningEffort: "high",
	})

	if _, err := c.Execute(context.Background(), UpdateAgentParams{ID: "agent-1", Name: ptrTo("renamed")}); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !repo.LastUpdate.ReasoningEffortEnabled || repo.LastUpdate.ReasoningEffort != "high" {
		t.Errorf("reasoning = (%v, %q), want (true, %q)", repo.LastUpdate.ReasoningEffortEnabled, repo.LastUpdate.ReasoningEffort, "high")
	}
}

func TestUpdateAgent_ReasoningEffort_CanBeSetAndDisabled(t *testing.T) {
	c, repo := newReasoningUpdateCommand(&agent.Agent{ID: "agent-1", ProviderID: "provider-1", ModelID: "o3"})

	_, err := c.Execute(context.Background(), UpdateAgentParams{
		ID: "agent-1", ReasoningEffortEnabled: ptrTo(true), ReasoningEffort: ptrTo("low"),
	})
	if err != nil {
		t.Fatalf("Execute (enable): %v", err)
	}
	if !repo.LastUpdate.ReasoningEffortEnabled || repo.LastUpdate.ReasoningEffort != "low" {
		t.Errorf("reasoning = (%v, %q), want (true, %q)", repo.LastUpdate.ReasoningEffortEnabled, repo.LastUpdate.ReasoningEffort, "low")
	}

	// Turning the option off keeps the stored value, so a user can switch it back on.
	_, err = c.Execute(context.Background(), UpdateAgentParams{
		ID: "agent-1", ReasoningEffortEnabled: ptrTo(false),
	})
	if err != nil {
		t.Fatalf("Execute (disable): %v", err)
	}
	if repo.LastUpdate.ReasoningEffortEnabled || repo.LastUpdate.ReasoningEffort != "low" {
		t.Errorf("reasoning = (%v, %q), want (false, %q)", repo.LastUpdate.ReasoningEffortEnabled, repo.LastUpdate.ReasoningEffort, "low")
	}
}

func TestUpdateAgent_ReasoningEffort_EnablingWithoutValueIsInvalid(t *testing.T) {
	c, repo := newReasoningUpdateCommand(&agent.Agent{ID: "agent-1", ProviderID: "provider-1", ModelID: "o3"})

	_, err := c.Execute(context.Background(), UpdateAgentParams{
		ID: "agent-1", ReasoningEffortEnabled: ptrTo(true),
	})
	assertInvalidInput(t, err)
	if repo.LastUpdate != nil {
		t.Error("agent must not be updated when the input is invalid")
	}
}
