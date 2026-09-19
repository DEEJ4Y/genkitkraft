package httphandler_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/DEEJ4Y/genkitkraft/internal/api/gen"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/playground"
)

func createPlaygroundSession(t *testing.T, env *testEnv) string {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/agents/"+env.agentID+"/playground/sessions",
		makeDeployRequest(t, map[string]interface{}{}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("create playground session: expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp gen.ModelsPlaygroundSessionResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("create playground session: decode: %v", err)
	}
	return resp.Id
}

// sseEvent is one parsed "id: N\ndata: ..." pair (or just "data: ..." if no
// id: line preceded it, as with the [DONE]/[ERROR] sentinels).
type sseEvent struct {
	ID    int
	HasID bool
	Data  string
}

func parseSSEEvents(body *bytes.Buffer) []sseEvent {
	scanner := bufio.NewScanner(body)
	var events []sseEvent
	pendingID := 0
	hasPendingID := false
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "id: "):
			if n, err := strconv.Atoi(strings.TrimPrefix(line, "id: ")); err == nil {
				pendingID = n
				hasPendingID = true
			}
		case strings.HasPrefix(line, "data: "):
			events = append(events, sseEvent{ID: pendingID, HasID: hasPendingID, Data: strings.TrimPrefix(line, "data: ")})
			hasPendingID = false
			pendingID = 0
		}
	}
	return events
}

func TestPlaygroundChat_Streaming_SequenceIDs(t *testing.T) {
	env := setupTestEnv(t)
	sessionID := createPlaygroundSession(t, env)

	env.mockChat.StreamTokens = []string{"Hello", " world"}
	env.mockChat.StreamError = nil

	reqBody := map[string]interface{}{
		"sessionId": sessionID,
		"content":   "hi",
		"stream":    true,
	}
	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/agents/"+env.agentID+"/playground/chat",
		makeDeployRequest(t, reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	events := parseSSEEvents(w.Body)
	if len(events) < 3 {
		t.Fatalf("expected at least 3 SSE events (2 tokens + [DONE]), got %d: %+v", len(events), events)
	}

	if !events[0].HasID || events[0].ID != 1 {
		t.Errorf("expected first event id 1, got %+v", events[0])
	}
	if !events[1].HasID || events[1].ID != 2 {
		t.Errorf("expected second event id 2, got %+v", events[1])
	}

	last := events[len(events)-1]
	if last.Data != "[DONE]" {
		t.Errorf("expected last event '[DONE]', got %q", last.Data)
	}

	joined := events[0].Data + events[1].Data
	if joined != "Hello world" {
		t.Errorf("expected streamed content 'Hello world', got %q", joined)
	}
}

func TestPlaygroundChat_Streaming_MidStreamError_PersistsPartialWithErrorStatus(t *testing.T) {
	env := setupTestEnv(t)
	sessionID := createPlaygroundSession(t, env)

	env.mockChat.StreamTokens = []string{"partial ", "content"}
	env.mockChat.StreamError = errors.New("provider exploded")

	reqBody := map[string]interface{}{
		"sessionId": sessionID,
		"content":   "hi",
		"stream":    true,
	}
	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/agents/"+env.agentID+"/playground/chat",
		makeDeployRequest(t, reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	events := parseSSEEvents(w.Body)
	last := events[len(events)-1]
	if !strings.HasPrefix(last.Data, "[ERROR]") {
		t.Errorf("expected last event to start with '[ERROR]', got %q", last.Data)
	}

	msg, err := env.playgroundRepo.GetLatestMessageBySession(context.Background(), sessionID)
	if err != nil {
		t.Fatalf("get latest message: %v", err)
	}
	if msg.Role != "assistant" {
		t.Errorf("expected role 'assistant', got %q", msg.Role)
	}
	if msg.Status != playground.MessageStatusError {
		t.Errorf("expected status %q, got %q", playground.MessageStatusError, msg.Status)
	}
	if msg.Content != "partial content" {
		t.Errorf("expected partial content 'partial content' to survive the error, got %q", msg.Content)
	}
}

func TestPlaygroundChat_Reconnect_ReplaysFromLastEventID(t *testing.T) {
	env := setupTestEnv(t)
	sessionID := createPlaygroundSession(t, env)

	env.mockChat.StreamTokens = []string{"one ", "two ", "three"}
	env.mockChat.StreamError = nil

	reqBody := map[string]interface{}{
		"sessionId": sessionID,
		"content":   "hi",
		"stream":    true,
	}
	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/agents/"+env.agentID+"/playground/chat",
		makeDeployRequest(t, reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Reconnect as if only the first token (seq 1) had been received.
	reconnectReq := httptest.NewRequest(http.MethodGet,
		"/api/v1/agents/"+env.agentID+"/playground/sessions/"+sessionID+"/stream", nil)
	reconnectReq.Header.Set("Last-Event-ID", "1")
	reconnectW := httptest.NewRecorder()
	env.mux.ServeHTTP(reconnectW, reconnectReq)

	if reconnectW.Code != http.StatusOK {
		t.Fatalf("reconnect: expected 200, got %d: %s", reconnectW.Code, reconnectW.Body.String())
	}

	events := parseSSEEvents(reconnectW.Body)
	if len(events) < 3 {
		t.Fatalf("expected 2 replayed tokens + [DONE], got %d: %+v", len(events), events)
	}
	if events[0].ID != 2 || events[0].Data != "two " {
		t.Errorf("expected first replayed event to be seq 2 'two ', got %+v", events[0])
	}
	if events[1].ID != 3 || events[1].Data != "three" {
		t.Errorf("expected second replayed event to be seq 3 'three', got %+v", events[1])
	}
	if events[len(events)-1].Data != "[DONE]" {
		t.Errorf("expected last replayed event '[DONE]', got %q", events[len(events)-1].Data)
	}
}

func TestPlaygroundChat_Cancel_StopsInFlightStream(t *testing.T) {
	env := setupTestEnv(t)
	sessionID := createPlaygroundSession(t, env)

	env.mockChat.StreamTokens = []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	env.mockChat.StreamDelay = 100 * time.Millisecond
	env.mockChat.StreamError = nil

	reqBody := map[string]interface{}{
		"sessionId": sessionID,
		"content":   "hi",
		"stream":    true,
	}
	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/agents/"+env.agentID+"/playground/chat",
		makeDeployRequest(t, reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		env.mux.ServeHTTP(w, req)
		close(done)
	}()

	// Let a couple of tokens through, then cancel mid-stream.
	time.Sleep(150 * time.Millisecond)

	cancelReq := httptest.NewRequest(http.MethodPost,
		"/api/v1/agents/"+env.agentID+"/playground/sessions/"+sessionID+"/stream/cancel", nil)
	cancelW := httptest.NewRecorder()
	env.mux.ServeHTTP(cancelW, cancelReq)
	if cancelW.Code != http.StatusNoContent {
		t.Fatalf("cancel: expected 204, got %d: %s", cancelW.Code, cancelW.Body.String())
	}

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for streaming request to finish after cancel")
	}

	events := parseSSEEvents(w.Body)
	if len(events) == 0 {
		t.Fatal("expected at least some content to have streamed before cancel took effect")
	}
	if len(events) >= len(env.mockChat.StreamTokens)+1 {
		t.Errorf("expected cancel to stop generation before all %d tokens were sent, got %d events", len(env.mockChat.StreamTokens), len(events))
	}

	msg, err := env.playgroundRepo.GetLatestMessageBySession(context.Background(), sessionID)
	if err != nil {
		t.Fatalf("get latest message: %v", err)
	}
	if msg.Status != playground.MessageStatusError {
		t.Errorf("expected message status %q after cancel, got %q", playground.MessageStatusError, msg.Status)
	}
}
