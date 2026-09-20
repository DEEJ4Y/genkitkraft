package httphandler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DEEJ4Y/genkitkraft/internal/api/gen"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/agent"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/gap"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/playground"
)

// createSecondAgent seeds a second agent (with its own gaps, scoped
// separately from env.agentID) using the same seeded provider.
func createSecondAgent(t *testing.T, env *testEnv) string {
	t.Helper()
	a := &agent.Agent{Name: "second-agent", ProviderID: env.providerID, ModelID: "gpt-4o"}
	if err := env.agentRepo.Create(context.Background(), a); err != nil {
		t.Fatalf("create second agent: %v", err)
	}
	return a.ID
}

func seedGap(t *testing.T, env *testEnv, agentID string, g *gap.Gap) *gap.Gap {
	t.Helper()
	g.AgentID = agentID
	if g.Category == "" {
		g.Category = gap.CategoryKnowledge
	}
	if g.Context == "" {
		g.Context = "what does the refund policy say?"
	}
	if g.Details == "" {
		g.Details = "no refund policy document is configured"
	}
	if err := env.gapRepo.Create(context.Background(), g); err != nil {
		t.Fatalf("seed gap: %v", err)
	}
	return g
}

func TestListGaps_Empty(t *testing.T) {
	env := setupTestEnv(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/agents/"+env.agentID+"/gaps", nil)
	w := httptest.NewRecorder()
	env.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp gen.ModelsGapListResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Gaps) != 0 || resp.Total != 0 {
		t.Errorf("expected empty list, got %+v", resp)
	}
	if resp.Limit != 20 || resp.Offset != 0 {
		t.Errorf("expected default limit=20 offset=0, got limit=%d offset=%d", resp.Limit, resp.Offset)
	}
}

func TestListGaps_ScopedToAgent(t *testing.T) {
	env := setupTestEnv(t)
	otherAgentID := createSecondAgent(t, env)

	seedGap(t, env, env.agentID, &gap.Gap{})
	seedGap(t, env, otherAgentID, &gap.Gap{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/agents/"+env.agentID+"/gaps", nil)
	w := httptest.NewRecorder()
	env.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp gen.ModelsGapListResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Total != 1 || len(resp.Gaps) != 1 {
		t.Fatalf("expected exactly 1 gap for this agent, got %+v", resp)
	}
	if resp.Gaps[0].AgentId != env.agentID {
		t.Errorf("gap.AgentId = %q, want %q — a gap from a different agent leaked into this list", resp.Gaps[0].AgentId, env.agentID)
	}
}

func TestGetGap_NotFound(t *testing.T) {
	env := setupTestEnv(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/agents/"+env.agentID+"/gaps/nonexistent-gap-id", nil)
	w := httptest.NewRecorder()
	env.mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

// A gap that exists but belongs to a different agent must 404, the same as
// one that doesn't exist at all — otherwise one agent's gaps would be
// readable through another agent's URL.
func TestGetGap_BelongsToDifferentAgent_Returns404(t *testing.T) {
	env := setupTestEnv(t)
	otherAgentID := createSecondAgent(t, env)
	g := seedGap(t, env, otherAgentID, &gap.Gap{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/agents/"+env.agentID+"/gaps/"+g.ID, nil)
	w := httptest.NewRecorder()
	env.mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for a gap owned by a different agent, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetGap_HappyPath_IncludesReferences(t *testing.T) {
	env := setupTestEnv(t)
	g := seedGap(t, env, env.agentID, &gap.Gap{SuggestedResolution: "add the refund policy as a source"})

	session := &playground.Session{AgentID: env.agentID}
	if err := env.playgroundRepo.CreateSession(context.Background(), session); err != nil {
		t.Fatalf("seed session: %v", err)
	}
	if err := env.gapRepo.AddReference(context.Background(), &gap.Reference{GapID: g.ID, SessionID: session.ID}); err != nil {
		t.Fatalf("seed reference: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/agents/"+env.agentID+"/gaps/"+g.ID, nil)
	w := httptest.NewRecorder()
	env.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp gen.ModelsGapResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Id != g.ID {
		t.Errorf("id = %q, want %q", resp.Id, g.ID)
	}
	if resp.SuggestedResolution == nil || *resp.SuggestedResolution != "add the refund policy as a source" {
		t.Errorf("suggestedResolution = %v, want the seeded value", resp.SuggestedResolution)
	}
	if len(resp.References) != 1 || resp.References[0].SessionId != session.ID {
		t.Errorf("references = %+v, want a reference to session %q", resp.References, session.ID)
	}
}

func TestUpdateGap_MissingStatus_Returns400(t *testing.T) {
	env := setupTestEnv(t)
	g := seedGap(t, env, env.agentID, &gap.Gap{})

	req := httptest.NewRequest(http.MethodPut, "/api/v1/agents/"+env.agentID+"/gaps/"+g.ID, makeDeployRequest(t, map[string]interface{}{}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestUpdateGap_NotFound_Returns404(t *testing.T) {
	env := setupTestEnv(t)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/agents/"+env.agentID+"/gaps/nonexistent-gap-id",
		makeDeployRequest(t, map[string]interface{}{"status": "resolved"}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestUpdateGap_BelongsToDifferentAgent_Returns404(t *testing.T) {
	env := setupTestEnv(t)
	otherAgentID := createSecondAgent(t, env)
	g := seedGap(t, env, otherAgentID, &gap.Gap{})

	req := httptest.NewRequest(http.MethodPut, "/api/v1/agents/"+env.agentID+"/gaps/"+g.ID,
		makeDeployRequest(t, map[string]interface{}{"status": "resolved"}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for a gap owned by a different agent, got %d: %s", w.Code, w.Body.String())
	}
}

func TestUpdateGap_Resolve(t *testing.T) {
	env := setupTestEnv(t)
	g := seedGap(t, env, env.agentID, &gap.Gap{})

	req := httptest.NewRequest(http.MethodPut, "/api/v1/agents/"+env.agentID+"/gaps/"+g.ID,
		makeDeployRequest(t, map[string]interface{}{"status": "resolved"}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp gen.ModelsGapResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Status != gen.Resolved {
		t.Errorf("status = %q, want %q", resp.Status, gen.Resolved)
	}
}

func TestUpdateGap_Dismiss_RequiresDismissalCategory(t *testing.T) {
	env := setupTestEnv(t)
	g := seedGap(t, env, env.agentID, &gap.Gap{})

	req := httptest.NewRequest(http.MethodPut, "/api/v1/agents/"+env.agentID+"/gaps/"+g.ID,
		makeDeployRequest(t, map[string]interface{}{"status": "dismissed"}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 without a dismissalCategory, got %d: %s", w.Code, w.Body.String())
	}
}

// Dismissing as "unrelated" is terminal — a second request trying to reopen
// it must be rejected (409), proving the terminal rule holds through the
// full HTTP stack, not just at the command layer.
func TestUpdateGap_DismissUnrelated_ThenReopenFails(t *testing.T) {
	env := setupTestEnv(t)
	g := seedGap(t, env, env.agentID, &gap.Gap{})

	dismissReq := httptest.NewRequest(http.MethodPut, "/api/v1/agents/"+env.agentID+"/gaps/"+g.ID,
		makeDeployRequest(t, map[string]interface{}{"status": "dismissed", "dismissalCategory": "unrelated"}))
	dismissReq.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	env.mux.ServeHTTP(w1, dismissReq)
	if w1.Code != http.StatusOK {
		t.Fatalf("dismiss: expected 200, got %d: %s", w1.Code, w1.Body.String())
	}

	reopenReq := httptest.NewRequest(http.MethodPut, "/api/v1/agents/"+env.agentID+"/gaps/"+g.ID,
		makeDeployRequest(t, map[string]interface{}{"status": "open"}))
	reopenReq.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	env.mux.ServeHTTP(w2, reopenReq)

	if w2.Code != http.StatusConflict {
		t.Fatalf("reopen after unrelated dismissal: expected 409, got %d: %s", w2.Code, w2.Body.String())
	}
}
