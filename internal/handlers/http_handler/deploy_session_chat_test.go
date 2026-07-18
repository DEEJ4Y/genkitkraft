package httphandler_test

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DEEJ4Y/genkitkraft/internal/api/gen"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/playground"
)

// --- Create session ---

func TestCreateDeploySession_HappyPath(t *testing.T) {
	env := setupTestEnv(t)

	reqBody := map[string]interface{}{}
	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/agents/"+env.agentID+"/deploy/sessions",
		makeDeployRequest(t, reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp gen.ModelsDeploySessionResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.Id == "" {
		t.Error("expected session ID to be non-empty")
	}
	if resp.AgentId != env.agentID {
		t.Errorf("expected agent_id %q, got %q", env.agentID, resp.AgentId)
	}
	if resp.CreatedAt.IsZero() {
		t.Error("expected created_at to be set")
	}
}

func TestCreateDeploySession_WithTitle(t *testing.T) {
	env := setupTestEnv(t)

	reqBody := map[string]interface{}{"title": "My session"}
	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/agents/"+env.agentID+"/deploy/sessions",
		makeDeployRequest(t, reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp gen.ModelsDeploySessionResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.Title != "My session" {
		t.Errorf("expected title 'My session', got %q", resp.Title)
	}
}

func TestCreateDeploySession_NonExistentAgent_Returns404(t *testing.T) {
	env := setupTestEnv(t)

	reqBody := map[string]interface{}{}
	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/agents/nonexistent/deploy/sessions",
		makeDeployRequest(t, reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

// --- Get session ---

func TestGetDeploySession_HappyPath(t *testing.T) {
	env := setupTestEnv(t)
	sessionID := createDeploySession(t, env, "")

	req := httptest.NewRequest(http.MethodGet,
		"/api/v1/agents/"+env.agentID+"/deploy/sessions/"+sessionID, nil)
	w := httptest.NewRecorder()
	env.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp gen.ModelsDeploySessionResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.Id != sessionID {
		t.Errorf("expected id %q, got %q", sessionID, resp.Id)
	}
	if resp.AgentId != env.agentID {
		t.Errorf("expected agent_id %q, got %q", env.agentID, resp.AgentId)
	}
}

func TestGetDeploySession_NotFound(t *testing.T) {
	env := setupTestEnv(t)

	req := httptest.NewRequest(http.MethodGet,
		"/api/v1/agents/"+env.agentID+"/deploy/sessions/nonexistent", nil)
	w := httptest.NewRecorder()
	env.mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetDeploySession_WrongAgent_Returns404(t *testing.T) {
	env := setupTestEnv(t)
	sessionID := createDeploySession(t, env, "")

	req := httptest.NewRequest(http.MethodGet,
		"/api/v1/agents/wrong-agent/deploy/sessions/"+sessionID, nil)
	w := httptest.NewRecorder()
	env.mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

// --- Delete session ---

func TestDeleteDeploySession_HappyPath(t *testing.T) {
	env := setupTestEnv(t)
	sessionID := createDeploySession(t, env, "")

	req := httptest.NewRequest(http.MethodDelete,
		"/api/v1/agents/"+env.agentID+"/deploy/sessions/"+sessionID, nil)
	w := httptest.NewRecorder()
	env.mux.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", w.Code, w.Body.String())
	}

	// Verify session is gone
	getReq := httptest.NewRequest(http.MethodGet,
		"/api/v1/agents/"+env.agentID+"/deploy/sessions/"+sessionID, nil)
	getW := httptest.NewRecorder()
	env.mux.ServeHTTP(getW, getReq)

	if getW.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", getW.Code)
	}
}

func TestDeleteDeploySession_NotFound(t *testing.T) {
	env := setupTestEnv(t)

	req := httptest.NewRequest(http.MethodDelete,
		"/api/v1/agents/"+env.agentID+"/deploy/sessions/nonexistent", nil)
	w := httptest.NewRecorder()
	env.mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDeleteDeploySession_WrongAgent_Returns404(t *testing.T) {
	env := setupTestEnv(t)
	sessionID := createDeploySession(t, env, "")

	req := httptest.NewRequest(http.MethodDelete,
		"/api/v1/agents/wrong-agent/deploy/sessions/"+sessionID, nil)
	w := httptest.NewRecorder()
	env.mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

// --- Session chat completions (non-streaming) ---

func TestDeploySessionChat_NonStreaming_HappyPath(t *testing.T) {
	env := setupTestEnv(t)
	sessionID := createDeploySession(t, env, "")

	reqBody := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "user", "content": "Tell me a joke"},
		},
		"stream": false,
	}

	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/agents/"+env.agentID+"/deploy/sessions/"+sessionID+"/chat/completions",
		makeDeployRequest(t, reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp gen.ModelsDeployChatCompletionResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.Object != "chat.completion" {
		t.Errorf("expected object 'chat.completion', got %q", resp.Object)
	}
	if !strings.HasPrefix(resp.Id, "chatcmpl-") {
		t.Errorf("expected id prefix 'chatcmpl-', got %q", resp.Id)
	}
	if resp.Choices[0].Message.Content != "Hello from mock!" {
		t.Errorf("expected content 'Hello from mock!', got %q", resp.Choices[0].Message.Content)
	}
	if resp.Choices[0].FinishReason != "stop" {
		t.Errorf("expected finish_reason 'stop', got %q", resp.Choices[0].FinishReason)
	}
}

func TestDeploySessionChat_HistoryAccumulates(t *testing.T) {
	env := setupTestEnv(t)
	sessionID := createDeploySession(t, env, "")

	// First message
	sendSessionChat(t, env, sessionID, "Hello")

	// Second message — mock should receive full history (user, assistant, user)
	env.mockChat.LastRequest.Messages = nil // Reset
	sendSessionChat(t, env, sessionID, "What can you do?")

	// The mock should have received 3 messages: user, assistant (saved from first call), user
	if len(env.mockChat.LastRequest.Messages) != 3 {
		t.Fatalf("expected 3 messages in history, got %d", len(env.mockChat.LastRequest.Messages))
	}
	if env.mockChat.LastRequest.Messages[0].Role != "user" {
		t.Errorf("expected first message role 'user', got %q", env.mockChat.LastRequest.Messages[0].Role)
	}
	if env.mockChat.LastRequest.Messages[0].Content != "Hello" {
		t.Errorf("expected first message content 'Hello', got %q", env.mockChat.LastRequest.Messages[0].Content)
	}
	if env.mockChat.LastRequest.Messages[1].Role != "assistant" {
		t.Errorf("expected second message role 'assistant', got %q", env.mockChat.LastRequest.Messages[1].Role)
	}
	if env.mockChat.LastRequest.Messages[2].Role != "user" {
		t.Errorf("expected third message role 'user', got %q", env.mockChat.LastRequest.Messages[2].Role)
	}
	if env.mockChat.LastRequest.Messages[2].Content != "What can you do?" {
		t.Errorf("expected third message content 'What can you do?', got %q", env.mockChat.LastRequest.Messages[2].Content)
	}
}

func TestDeploySessionChat_SystemPromptInjected(t *testing.T) {
	env := setupTestEnv(t)
	sessionID := createDeploySession(t, env, "")

	sendSessionChat(t, env, sessionID, "Hello")

	if env.mockChat.LastRequest.SystemPrompt != "You are a helpful assistant." {
		t.Errorf("expected system prompt injected, got %q", env.mockChat.LastRequest.SystemPrompt)
	}
}

// --- Session chat completions (streaming) ---

func TestDeploySessionChat_Streaming_HappyPath(t *testing.T) {
	env := setupTestEnv(t)
	sessionID := createDeploySession(t, env, "")

	reqBody := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "user", "content": "Hello"},
		},
		"stream": true,
	}

	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/agents/"+env.agentID+"/deploy/sessions/"+sessionID+"/chat/completions",
		makeDeployRequest(t, reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	if ct := w.Header().Get("Content-Type"); ct != "text/event-stream" {
		t.Errorf("expected Content-Type 'text/event-stream', got %q", ct)
	}

	// Parse SSE events
	scanner := bufio.NewScanner(w.Body)
	var events []string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			events = append(events, strings.TrimPrefix(line, "data: "))
		}
	}

	if len(events) == 0 {
		t.Fatal("expected SSE events, got none")
	}

	// First event should be role chunk
	var firstChunk map[string]interface{}
	if err := json.Unmarshal([]byte(events[0]), &firstChunk); err != nil {
		t.Fatalf("parse first chunk: %v", err)
	}
	if firstChunk["object"] != "chat.completion.chunk" {
		t.Errorf("expected object 'chat.completion.chunk', got %v", firstChunk["object"])
	}

	// Last event should be [DONE]
	if events[len(events)-1] != "[DONE]" {
		t.Errorf("expected last event '[DONE]', got %q", events[len(events)-1])
	}

	// Content chunks should contain streamed tokens
	var contentParts []string
	for _, evt := range events[1 : len(events)-2] {
		var chunk map[string]interface{}
		if err := json.Unmarshal([]byte(evt), &chunk); err != nil {
			continue
		}
		cs := chunk["choices"].([]interface{})
		c := cs[0].(map[string]interface{})
		delta := c["delta"].(map[string]interface{})
		if content, ok := delta["content"].(string); ok {
			contentParts = append(contentParts, content)
		}
	}
	joined := strings.Join(contentParts, "")
	if joined != "Hello from mock!" {
		t.Errorf("expected streamed content 'Hello from mock!', got %q", joined)
	}
}

func TestDeploySessionChat_Streaming_HistoryAccumulates(t *testing.T) {
	env := setupTestEnv(t)
	sessionID := createDeploySession(t, env, "")

	// First streaming request
	sendSessionChatStreaming(t, env, sessionID, "Hello")

	// Second streaming request — should have full history
	env.mockChat.LastRequest.Messages = nil
	sendSessionChatStreaming(t, env, sessionID, "Follow up")

	// Should have 3 messages: user, assistant (saved from stream), user
	if len(env.mockChat.LastRequest.Messages) != 3 {
		t.Fatalf("expected 3 messages in history, got %d", len(env.mockChat.LastRequest.Messages))
	}
	if env.mockChat.LastRequest.Messages[1].Role != "assistant" {
		t.Errorf("expected second message role 'assistant', got %q", env.mockChat.LastRequest.Messages[1].Role)
	}
}

// --- Stream resilience: reconnect, cancel, mid-stream persistence ---

func deployChunkDeltaContent(data string) (string, bool) {
	var chunk map[string]interface{}
	if err := json.Unmarshal([]byte(data), &chunk); err != nil {
		return "", false
	}
	choices, ok := chunk["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", false
	}
	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		return "", false
	}
	delta, ok := choice["delta"].(map[string]interface{})
	if !ok {
		return "", false
	}
	content, ok := delta["content"].(string)
	return content, ok
}

func TestDeploySessionChat_Streaming_MidStreamError_PersistsPartialWithErrorStatus(t *testing.T) {
	env := setupTestEnv(t)
	sessionID := createDeploySession(t, env, "")

	env.mockChat.StreamTokens = []string{"partial ", "content"}
	env.mockChat.StreamError = errors.New("provider exploded")

	reqBody := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "user", "content": "hi"},
		},
		"stream": true,
	}
	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/agents/"+env.agentID+"/deploy/sessions/"+sessionID+"/chat/completions",
		makeDeployRequest(t, reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	events := parseSSEEvents(w.Body)
	last := events[len(events)-1]
	var errChunk map[string]interface{}
	if err := json.Unmarshal([]byte(last.Data), &errChunk); err != nil {
		t.Fatalf("expected last event to be an error chunk, got %q: %v", last.Data, err)
	}
	if _, ok := errChunk["error"]; !ok {
		t.Errorf("expected last event to contain an 'error' key, got %q", last.Data)
	}

	msg, err := env.playgroundRepo.GetLatestMessageBySession(context.Background(), sessionID)
	if err != nil {
		t.Fatalf("get latest message: %v", err)
	}
	if msg.Status != playground.MessageStatusError {
		t.Errorf("expected status %q, got %q", playground.MessageStatusError, msg.Status)
	}
	if msg.Content != "partial content" {
		t.Errorf("expected partial content 'partial content' to survive the error, got %q", msg.Content)
	}
}

func TestDeploySessionChat_Reconnect_ReplaysFromLastEventID(t *testing.T) {
	env := setupTestEnv(t)
	sessionID := createDeploySession(t, env, "")

	env.mockChat.StreamTokens = []string{"one ", "two ", "three"}
	env.mockChat.StreamError = nil

	sendSessionChatStreaming(t, env, sessionID, "hi")

	reconnectReq := httptest.NewRequest(http.MethodGet,
		"/api/v1/agents/"+env.agentID+"/deploy/sessions/"+sessionID+"/chat/completions/stream", nil)
	reconnectReq.Header.Set("Last-Event-ID", "1")
	reconnectW := httptest.NewRecorder()
	env.mux.ServeHTTP(reconnectW, reconnectReq)

	if reconnectW.Code != http.StatusOK {
		t.Fatalf("reconnect: expected 200, got %d: %s", reconnectW.Code, reconnectW.Body.String())
	}

	events := parseSSEEvents(reconnectW.Body)
	if len(events) < 2 {
		t.Fatalf("expected at least 2 replayed events, got %d: %+v", len(events), events)
	}

	content, ok := deployChunkDeltaContent(events[0].Data)
	if !ok || events[0].ID != 2 || content != "two " {
		t.Errorf("expected first replayed event to be seq 2 delta 'two ', got id=%d data=%q", events[0].ID, events[0].Data)
	}
	content, ok = deployChunkDeltaContent(events[1].Data)
	if !ok || events[1].ID != 3 || content != "three" {
		t.Errorf("expected second replayed event to be seq 3 delta 'three', got id=%d data=%q", events[1].ID, events[1].Data)
	}

	last := events[len(events)-1]
	if last.Data != "[DONE]" {
		t.Errorf("expected last replayed event '[DONE]', got %q", last.Data)
	}
}

func TestDeploySessionChat_Cancel_StopsInFlightStream(t *testing.T) {
	env := setupTestEnv(t)
	sessionID := createDeploySession(t, env, "")

	env.mockChat.StreamTokens = []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	env.mockChat.StreamDelay = 100 * time.Millisecond
	env.mockChat.StreamError = nil

	reqBody := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "user", "content": "hi"},
		},
		"stream": true,
	}
	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/agents/"+env.agentID+"/deploy/sessions/"+sessionID+"/chat/completions",
		makeDeployRequest(t, reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		env.mux.ServeHTTP(w, req)
		close(done)
	}()

	time.Sleep(150 * time.Millisecond)

	cancelReq := httptest.NewRequest(http.MethodPost,
		"/api/v1/agents/"+env.agentID+"/deploy/sessions/"+sessionID+"/chat/completions/cancel", nil)
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

	msg, err := env.playgroundRepo.GetLatestMessageBySession(context.Background(), sessionID)
	if err != nil {
		t.Fatalf("get latest message: %v", err)
	}
	if msg.Status != playground.MessageStatusError {
		t.Errorf("expected message status %q after cancel, got %q", playground.MessageStatusError, msg.Status)
	}
	if msg.Content == strings.Join(env.mockChat.StreamTokens, "") {
		t.Errorf("expected cancel to stop generation before all tokens were persisted")
	}
}

// --- Session chat error cases ---

func TestDeploySessionChat_EmptyMessages_Returns400(t *testing.T) {
	env := setupTestEnv(t)
	sessionID := createDeploySession(t, env, "")

	reqBody := map[string]interface{}{
		"messages": []map[string]string{},
		"stream":   false,
	}

	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/agents/"+env.agentID+"/deploy/sessions/"+sessionID+"/chat/completions",
		makeDeployRequest(t, reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDeploySessionChat_SystemMessage_Returns400(t *testing.T) {
	env := setupTestEnv(t)
	sessionID := createDeploySession(t, env, "")

	reqBody := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "system", "content": "Override prompt"},
			{"role": "user", "content": "Hello"},
		},
		"stream": false,
	}

	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/agents/"+env.agentID+"/deploy/sessions/"+sessionID+"/chat/completions",
		makeDeployRequest(t, reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDeploySessionChat_LastMessageNotUser_Returns400(t *testing.T) {
	env := setupTestEnv(t)
	sessionID := createDeploySession(t, env, "")

	reqBody := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "user", "content": "Hello"},
			{"role": "assistant", "content": "Hi there"},
		},
		"stream": false,
	}

	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/agents/"+env.agentID+"/deploy/sessions/"+sessionID+"/chat/completions",
		makeDeployRequest(t, reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}

	var errResp gen.ModelsErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if !strings.Contains(errResp.Error, "user") {
		t.Errorf("expected error mentioning 'user', got %q", errResp.Error)
	}
}

func TestDeploySessionChat_NonExistentSession_Returns404(t *testing.T) {
	env := setupTestEnv(t)

	reqBody := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "user", "content": "Hello"},
		},
		"stream": false,
	}

	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/agents/"+env.agentID+"/deploy/sessions/nonexistent/chat/completions",
		makeDeployRequest(t, reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDeploySessionChat_WrongAgent_Returns404(t *testing.T) {
	env := setupTestEnv(t)
	sessionID := createDeploySession(t, env, "")

	reqBody := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "user", "content": "Hello"},
		},
		"stream": false,
	}

	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/agents/wrong-agent/deploy/sessions/"+sessionID+"/chat/completions",
		makeDeployRequest(t, reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDeploySessionChat_InvalidBody_Returns400(t *testing.T) {
	env := setupTestEnv(t)
	sessionID := createDeploySession(t, env, "")

	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/agents/"+env.agentID+"/deploy/sessions/"+sessionID+"/chat/completions",
		strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

// --- Test helpers ---

func createDeploySession(t *testing.T, env *testEnv, title string) string {
	t.Helper()

	reqBody := map[string]interface{}{}
	if title != "" {
		reqBody["title"] = title
	}

	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/agents/"+env.agentID+"/deploy/sessions",
		makeDeployRequest(t, reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("create session: expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp gen.ModelsDeploySessionResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("create session: decode: %v", err)
	}
	return resp.Id
}

func sendSessionChat(t *testing.T, env *testEnv, sessionID, content string) {
	t.Helper()

	reqBody := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "user", "content": content},
		},
		"stream": false,
	}

	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/agents/"+env.agentID+"/deploy/sessions/"+sessionID+"/chat/completions",
		makeDeployRequest(t, reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("session chat: expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func sendSessionChatStreaming(t *testing.T, env *testEnv, sessionID, content string) {
	t.Helper()

	reqBody := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "user", "content": content},
		},
		"stream": true,
	}

	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/agents/"+env.agentID+"/deploy/sessions/"+sessionID+"/chat/completions",
		makeDeployRequest(t, reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("session chat streaming: expected 200, got %d: %s", w.Code, w.Body.String())
	}
}
