package httphandler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DEEJ4Y/genkitkraft/internal/api/gen"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/agent"
)

// getDeploy issues a GET against the test mux and returns the recorder.
func getDeploy(t *testing.T, env *testEnv, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	w := httptest.NewRecorder()
	env.mux.ServeHTTP(w, req)
	return w
}

// createOtherAgent seeds a second agent (same provider) so tests can check agent scoping.
func createOtherAgent(t *testing.T, env *testEnv) string {
	t.Helper()
	a := &agent.Agent{Name: "other-agent", ProviderID: env.providerID, ModelID: "gpt-4o"}
	if err := env.agentRepo.Create(context.Background(), a); err != nil {
		t.Fatalf("create other agent: %v", err)
	}
	return a.ID
}

func decodeMessages(t *testing.T, w *httptest.ResponseRecorder) gen.ModelsDeployMessageListResponse {
	t.Helper()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp gen.ModelsDeployMessageListResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode messages: %v", err)
	}
	return resp
}

func decodeSessionList(t *testing.T, w *httptest.ResponseRecorder) gen.ModelsDeploySessionListResponse {
	t.Helper()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp gen.ModelsDeploySessionListResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode session list: %v", err)
	}
	return resp
}

// --- GET /deploy/sessions/{sessionId}/messages ---

func TestListDeployMessages_HappyPath(t *testing.T) {
	env := setupTestEnv(t)
	sessionID := createDeploySession(t, env, "")
	sendSessionChat(t, env, sessionID, "Hello")
	sendSessionChat(t, env, sessionID, "And again")

	w := getDeploy(t, env, "/api/v1/agents/"+env.agentID+"/deploy/sessions/"+sessionID+"/messages")
	resp := decodeMessages(t, w)

	want := []struct{ role, content string }{
		{"user", "Hello"},
		{"assistant", "Hello from mock!"},
		{"user", "And again"},
		{"assistant", "Hello from mock!"},
	}
	if len(resp.Messages) != len(want) {
		t.Fatalf("expected %d messages, got %d: %s", len(want), len(resp.Messages), w.Body.String())
	}
	for i, m := range resp.Messages {
		if m.Role != want[i].role || m.Content != want[i].content {
			t.Errorf("message %d: expected %s/%q, got %s/%q", i, want[i].role, want[i].content, m.Role, m.Content)
		}
		if m.Status != gen.Complete {
			t.Errorf("message %d: expected status complete, got %q", i, m.Status)
		}
		if m.SessionId != sessionID {
			t.Errorf("message %d: expected session_id %q, got %q", i, sessionID, m.SessionId)
		}
		if m.Id == "" || m.CreatedAt.IsZero() {
			t.Errorf("message %d: expected id and created_at to be set", i)
		}
	}
}

func TestListDeployMessages_UsesSnakeCaseFields(t *testing.T) {
	env := setupTestEnv(t)
	sessionID := createDeploySession(t, env, "")
	sendSessionChat(t, env, sessionID, "Hello")

	w := getDeploy(t, env, "/api/v1/agents/"+env.agentID+"/deploy/sessions/"+sessionID+"/messages")
	var raw struct {
		Messages []map[string]any `json:"messages"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &raw); err != nil {
		t.Fatalf("decode: %v", err)
	}
	for _, key := range []string{"id", "session_id", "role", "content", "status", "created_at"} {
		if _, ok := raw.Messages[0][key]; !ok {
			t.Errorf("expected key %q in message, got %v", key, raw.Messages[0])
		}
	}
}

func TestListDeployMessages_EmptySession_ReturnsEmptyArray(t *testing.T) {
	env := setupTestEnv(t)
	sessionID := createDeploySession(t, env, "")

	w := getDeploy(t, env, "/api/v1/agents/"+env.agentID+"/deploy/sessions/"+sessionID+"/messages")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"messages":[]`) {
		t.Errorf("expected an empty array, not null, got %s", w.Body.String())
	}
}

func TestListDeployMessages_NonExistentSession_Returns404(t *testing.T) {
	env := setupTestEnv(t)

	w := getDeploy(t, env, "/api/v1/agents/"+env.agentID+"/deploy/sessions/does-not-exist/messages")
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestListDeployMessages_WrongAgent_Returns404(t *testing.T) {
	env := setupTestEnv(t)
	sessionID := createDeploySession(t, env, "")
	sendSessionChat(t, env, sessionID, "secret")
	otherAgentID := createOtherAgent(t, env)

	w := getDeploy(t, env, "/api/v1/agents/"+otherAgentID+"/deploy/sessions/"+sessionID+"/messages")
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 when reading another agent's session, got %d: %s", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), "secret") {
		t.Errorf("response must not leak message content: %s", w.Body.String())
	}
}

func TestListDeployMessages_FailedReply_ExposesPartialContentWithErrorStatus(t *testing.T) {
	env := setupTestEnv(t)
	sessionID := createDeploySession(t, env, "")
	env.mockChat.StreamTokens = []string{"partial ", "content"}
	env.mockChat.StreamError = errors.New("provider exploded")
	sendSessionChatStreamingAllowError(t, env, sessionID, "hi")

	w := getDeploy(t, env, "/api/v1/agents/"+env.agentID+"/deploy/sessions/"+sessionID+"/messages")
	resp := decodeMessages(t, w)

	if len(resp.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d: %s", len(resp.Messages), w.Body.String())
	}
	last := resp.Messages[1]
	if last.Role != "assistant" || last.Status != gen.Error || last.Content != "partial content" {
		t.Errorf("expected assistant/error/%q, got %s/%s/%q", "partial content", last.Role, last.Status, last.Content)
	}
}

// sendSessionChatStreamingAllowError streams a chat turn and tolerates the mid-stream error
// (the response is still a 200 SSE stream that ends with an error chunk).
func sendSessionChatStreamingAllowError(t *testing.T, env *testEnv, sessionID, content string) {
	t.Helper()
	body := makeDeployRequest(t, map[string]interface{}{
		"messages": []map[string]string{{"role": "user", "content": content}},
		"stream":   true,
	})
	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/agents/"+env.agentID+"/deploy/sessions/"+sessionID+"/chat/completions", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("streaming chat: expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// --- GET /deploy/sessions ---

func TestListDeploySessions_HappyPath(t *testing.T) {
	env := setupTestEnv(t)
	first := createDeploySession(t, env, "first")
	createDeploySession(t, env, "second")

	w := getDeploy(t, env, "/api/v1/agents/"+env.agentID+"/deploy/sessions")
	resp := decodeSessionList(t, w)

	if resp.Total != 2 || len(resp.Sessions) != 2 {
		t.Fatalf("expected 2 sessions (total 2), got %d (total %d): %s", len(resp.Sessions), resp.Total, w.Body.String())
	}
	if resp.Limit != 20 || resp.Offset != 0 {
		t.Errorf("expected default limit 20 / offset 0, got %d / %d", resp.Limit, resp.Offset)
	}
	var found bool
	for _, s := range resp.Sessions {
		if s.Id == first {
			found = true
			if s.AgentId != env.agentID || s.Title != "first" || s.CreatedAt.IsZero() {
				t.Errorf("unexpected session fields: %+v", s)
			}
		}
	}
	if !found {
		t.Errorf("session %q missing from list: %s", first, w.Body.String())
	}
}

func TestListDeploySessions_MostRecentlyUpdatedFirst(t *testing.T) {
	env := setupTestEnv(t)
	older := createDeploySession(t, env, "older")
	time.Sleep(1100 * time.Millisecond) // timestamps may have one-second resolution
	newer := createDeploySession(t, env, "newer")

	resp := decodeSessionList(t, getDeploy(t, env, "/api/v1/agents/"+env.agentID+"/deploy/sessions"))
	if resp.Sessions[0].Id != newer || resp.Sessions[1].Id != older {
		t.Errorf("expected newest first [%s, %s], got [%s, %s]", newer, older, resp.Sessions[0].Id, resp.Sessions[1].Id)
	}

	time.Sleep(1100 * time.Millisecond)
	sendSessionChat(t, env, older, "bump") // activity moves a session to the front
	resp = decodeSessionList(t, getDeploy(t, env, "/api/v1/agents/"+env.agentID+"/deploy/sessions"))
	if resp.Sessions[0].Id != older {
		t.Errorf("expected the session with new activity first, got %s", resp.Sessions[0].Id)
	}
}

func TestListDeploySessions_Pagination(t *testing.T) {
	env := setupTestEnv(t)
	for i := 0; i < 3; i++ {
		createDeploySession(t, env, "s")
	}
	base := "/api/v1/agents/" + env.agentID + "/deploy/sessions"

	page1 := decodeSessionList(t, getDeploy(t, env, base+"?limit=2"))
	if len(page1.Sessions) != 2 || page1.Total != 3 || page1.Limit != 2 || page1.Offset != 0 {
		t.Errorf("page 1: got %d sessions, total %d, limit %d, offset %d", len(page1.Sessions), page1.Total, page1.Limit, page1.Offset)
	}

	page2 := decodeSessionList(t, getDeploy(t, env, base+"?limit=2&offset=2"))
	if len(page2.Sessions) != 1 || page2.Offset != 2 {
		t.Errorf("page 2: got %d sessions, offset %d", len(page2.Sessions), page2.Offset)
	}
	for _, s2 := range page2.Sessions {
		for _, s1 := range page1.Sessions {
			if s1.Id == s2.Id {
				t.Errorf("session %q appears on both pages", s1.Id)
			}
		}
	}

	w := getDeploy(t, env, base+"?offset=50")
	past := decodeSessionList(t, w)
	if len(past.Sessions) != 0 || past.Total != 3 {
		t.Errorf("offset past the end: expected no sessions and total 3, got %d / %d", len(past.Sessions), past.Total)
	}
	if !strings.Contains(w.Body.String(), `"sessions":[]`) {
		t.Errorf("expected an empty array, not null, got %s", w.Body.String())
	}
}

func TestListDeploySessions_ClampsLimitAndOffset(t *testing.T) {
	env := setupTestEnv(t)
	createDeploySession(t, env, "s")
	base := "/api/v1/agents/" + env.agentID + "/deploy/sessions"

	if r := decodeSessionList(t, getDeploy(t, env, base+"?limit=1000")); r.Limit != 100 {
		t.Errorf("expected limit clamped to 100, got %d", r.Limit)
	}
	if r := decodeSessionList(t, getDeploy(t, env, base+"?limit=0")); r.Limit != 20 {
		t.Errorf("expected limit 0 to fall back to 20, got %d", r.Limit)
	}
	if r := decodeSessionList(t, getDeploy(t, env, base+"?offset=-5")); r.Offset != 0 {
		t.Errorf("expected negative offset clamped to 0, got %d", r.Offset)
	}
}

func TestListDeploySessions_InvalidPagination_Returns400(t *testing.T) {
	env := setupTestEnv(t)

	w := getDeploy(t, env, "/api/v1/agents/"+env.agentID+"/deploy/sessions?limit=abc")
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestListDeploySessions_Empty_ReturnsEmptyArray(t *testing.T) {
	env := setupTestEnv(t)

	w := getDeploy(t, env, "/api/v1/agents/"+env.agentID+"/deploy/sessions")
	resp := decodeSessionList(t, w)
	if resp.Total != 0 || len(resp.Sessions) != 0 {
		t.Errorf("expected an empty list, got %d (total %d)", len(resp.Sessions), resp.Total)
	}
	if !strings.Contains(w.Body.String(), `"sessions":[]`) {
		t.Errorf("expected an empty array, not null, got %s", w.Body.String())
	}
}

func TestListDeploySessions_NonExistentAgent_Returns404(t *testing.T) {
	env := setupTestEnv(t)

	w := getDeploy(t, env, "/api/v1/agents/does-not-exist/deploy/sessions")
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestListDeploySessions_OnlyIncludesThatAgentsSessions(t *testing.T) {
	env := setupTestEnv(t)
	mine := createDeploySession(t, env, "mine")
	otherAgentID := createOtherAgent(t, env)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents/"+otherAgentID+"/deploy/sessions", makeDeployRequest(t, map[string]interface{}{"title": "theirs"}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.mux.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("create other agent session: expected 201, got %d: %s", w.Code, w.Body.String())
	}

	resp := decodeSessionList(t, getDeploy(t, env, "/api/v1/agents/"+env.agentID+"/deploy/sessions"))
	if resp.Total != 1 || len(resp.Sessions) != 1 || resp.Sessions[0].Id != mine {
		t.Errorf("expected only this agent's session %q, got %+v", mine, resp)
	}
}
