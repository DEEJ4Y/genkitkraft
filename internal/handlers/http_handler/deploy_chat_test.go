package httphandler_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rs/zerolog"

	aesgcmencryptor "github.com/DEEJ4Y/genkitkraft/internal/adapters/aesgcm_encryptor"
	inmemorystreamregistry "github.com/DEEJ4Y/genkitkraft/internal/adapters/in_memory_stream_registry"
	sqliteagent "github.com/DEEJ4Y/genkitkraft/internal/adapters/sqlite_agent"
	sqliteagenttool "github.com/DEEJ4Y/genkitkraft/internal/adapters/sqlite_agent_tool"
	sqlitedb "github.com/DEEJ4Y/genkitkraft/internal/adapters/sqlite_db"
	sqlitehttptool "github.com/DEEJ4Y/genkitkraft/internal/adapters/sqlite_http_tool"
	sqlitemcpserver "github.com/DEEJ4Y/genkitkraft/internal/adapters/sqlite_mcp_server"
	sqliteplayground "github.com/DEEJ4Y/genkitkraft/internal/adapters/sqlite_playground"
	sqliteprompt "github.com/DEEJ4Y/genkitkraft/internal/adapters/sqlite_prompt"
	sqliteprovider "github.com/DEEJ4Y/genkitkraft/internal/adapters/sqlite_provider"
	"github.com/DEEJ4Y/genkitkraft/internal/api/gen"
	"github.com/DEEJ4Y/genkitkraft/internal/app"
	"github.com/DEEJ4Y/genkitkraft/internal/app/commands"
	"github.com/DEEJ4Y/genkitkraft/internal/app/queries"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/agent"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/prompt"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/provider"
	httphandler "github.com/DEEJ4Y/genkitkraft/internal/handlers/http_handler"
	playgroundrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/playground_repo"
	mockchat "github.com/DEEJ4Y/genkitkraft/resources/test/mock"
)

// testEnv holds the wired-up test dependencies.
type testEnv struct {
	handler        *httphandler.Handler
	mux            *http.ServeMux
	agentID        string
	mockChat       *mockchat.ChatProvider
	playgroundRepo playgroundrepo.PlaygroundRepository
}

// setupTestEnv creates a fully wired test environment with a real SQLite DB,
// real adapters, and a mock ChatProvider. It seeds a provider, prompt, and agent.
func setupTestEnv(t *testing.T) *testEnv {
	t.Helper()

	// Open a temp SQLite DB (auto-runs migrations)
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := sqlitedb.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	// Real adapters
	enc, err := aesgcmencryptor.NewAESGCMEncryptor("test-encryption-key")
	if err != nil {
		t.Fatalf("create encryptor: %v", err)
	}
	agentRepo := sqliteagent.NewAgentRepository(db)
	providerRepo := sqliteprovider.NewProviderRepository(db)
	promptRepo := sqliteprompt.NewPromptRepository(db)
	playgroundRepo := sqliteplayground.NewPlaygroundRepository(db)
	agentToolRepo := sqliteagenttool.NewRepository(db)
	httpToolRepo := sqlitehttptool.NewHttpToolRepository(db)
	mcpServerRepo := sqlitemcpserver.NewMcpServerRepository(db)

	ctx := context.Background()

	// Seed provider with encrypted API key
	apiKey, err := enc.Encrypt("test-api-key")
	if err != nil {
		t.Fatalf("encrypt api key: %v", err)
	}
	p := &provider.Provider{
		Name:         "test-provider",
		ProviderType: provider.OpenAI,
		APIKey:       &apiKey,
		BaseURL:      "https://api.openai.com/v1",
		Enabled:      true,
	}
	if err := providerRepo.Create(ctx, p); err != nil {
		t.Fatalf("create provider: %v", err)
	}

	// Seed prompt
	pr := &prompt.Prompt{
		Name:    "test-prompt",
		Content: "You are a helpful assistant.",
	}
	if err := promptRepo.Create(ctx, pr); err != nil {
		t.Fatalf("create prompt: %v", err)
	}

	// Seed agent
	a := &agent.Agent{
		Name:               "test-agent",
		ProviderID:         p.ID,
		ModelID:            "gpt-4o",
		SystemPromptID:     pr.ID,
		TemperatureEnabled: true,
		Temperature:        0.7,
	}
	if err := agentRepo.Create(ctx, a); err != nil {
		t.Fatalf("create agent: %v", err)
	}

	mockCP := &mockchat.ChatProvider{
		ChatResponse: "Hello from mock!",
		StreamTokens: []string{"Hello", " from", " mock", "!"},
	}

	// Build app layer (minimal — only what deploy handler needs)
	resolveConfig := queries.NewResolvePlaygroundConfigQuery(agentRepo, providerRepo, promptRepo, enc, agentToolRepo, httpToolRepo, mcpServerRepo)
	saveMessage := commands.NewSavePlaygroundMessageCommand(playgroundRepo)
	createSession := commands.NewCreatePlaygroundSessionCommand(playgroundRepo, agentRepo)
	deleteSession := commands.NewDeletePlaygroundSessionCommand(playgroundRepo)
	streamRegistry := inmemorystreamregistry.NewRegistry()
	startStream := commands.NewStartPlaygroundStreamCommand(playgroundRepo, mockCP, streamRegistry, zerolog.Nop())
	cancelStream := commands.NewCancelPlaygroundStreamCommand(playgroundRepo, streamRegistry)
	failStream := commands.NewFailPlaygroundStreamCommand(playgroundRepo)
	listSessions := queries.NewListPlaygroundSessionsQuery(playgroundRepo)
	getSession := queries.NewGetPlaygroundSessionQuery(playgroundRepo)
	listMessages := queries.NewListPlaygroundMessagesQuery(playgroundRepo)
	getStreamChunks := queries.NewGetPlaygroundStreamChunksQuery(playgroundRepo)

	playgroundApp := &app.PlaygroundApp{
		Commands: app.PlaygroundCommands{
			CreateSession: createSession,
			DeleteSession: deleteSession,
			SaveMessage:   saveMessage,
			StartStream:   startStream,
			CancelStream:  cancelStream,
			FailStream:    failStream,
		},
		Queries: app.PlaygroundQueries{
			ListSessions:    listSessions,
			GetSession:      getSession,
			ListMessages:    listMessages,
			ResolveConfig:   resolveConfig,
			GetStreamChunks: getStreamChunks,
		},
	}

	handler := httphandler.NewHandler(nil, nil, nil, nil, playgroundApp, nil, nil, nil, nil, mockCP, nil)

	mux := http.NewServeMux()
	gen.HandlerFromMux(handler, mux)

	return &testEnv{
		handler:        handler,
		mux:            mux,
		agentID:        a.ID,
		mockChat:       mockCP,
		playgroundRepo: playgroundRepo,
	}
}

func makeDeployRequest(t *testing.T, body interface{}) *bytes.Buffer {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	return bytes.NewBuffer(b)
}

// --- Non-streaming tests ---

func TestDeployChat_NonStreaming_HappyPath(t *testing.T) {
	env := setupTestEnv(t)

	reqBody := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "user", "content": "Hello"},
		},
		"stream": false,
	}

	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/agents/"+env.agentID+"/deploy/chat/completions",
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
	if resp.Model != "gpt-4o" {
		t.Errorf("expected model 'gpt-4o', got %q", resp.Model)
	}
	if len(resp.Choices) != 1 {
		t.Fatalf("expected 1 choice, got %d", len(resp.Choices))
	}
	if resp.Choices[0].Message.Content != "Hello from mock!" {
		t.Errorf("expected content 'Hello from mock!', got %q", resp.Choices[0].Message.Content)
	}
	if resp.Choices[0].FinishReason != "stop" {
		t.Errorf("expected finish_reason 'stop', got %q", resp.Choices[0].FinishReason)
	}
	if string(resp.Choices[0].Message.Role) != "assistant" {
		t.Errorf("expected role 'assistant', got %q", resp.Choices[0].Message.Role)
	}
}

func TestDeployChat_DefaultStreamFalse(t *testing.T) {
	env := setupTestEnv(t)

	// Omit "stream" field entirely — should default to non-streaming
	reqBody := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "user", "content": "Hi"},
		},
	}

	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/agents/"+env.agentID+"/deploy/chat/completions",
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
		t.Errorf("expected non-streaming response, got object=%q", resp.Object)
	}
}

func TestDeployChat_MultipleMessages(t *testing.T) {
	env := setupTestEnv(t)

	reqBody := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "user", "content": "Hello"},
			{"role": "assistant", "content": "Hi there!"},
			{"role": "user", "content": "What can you do?"},
		},
		"stream": false,
	}

	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/agents/"+env.agentID+"/deploy/chat/completions",
		makeDeployRequest(t, reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify mock received all 3 messages
	if len(env.mockChat.LastRequest.Messages) != 3 {
		t.Errorf("expected 3 messages passed to provider, got %d", len(env.mockChat.LastRequest.Messages))
	}

	// Verify system prompt was injected
	if env.mockChat.LastRequest.SystemPrompt != "You are a helpful assistant." {
		t.Errorf("expected system prompt to be injected, got %q", env.mockChat.LastRequest.SystemPrompt)
	}
}

// --- Streaming tests ---

func TestDeployChat_Streaming_HappyPath(t *testing.T) {
	env := setupTestEnv(t)

	reqBody := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "user", "content": "Hello"},
		},
		"stream": true,
	}

	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/agents/"+env.agentID+"/deploy/chat/completions",
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
	lastEvent := events[len(events)-1]
	if lastEvent != "[DONE]" {
		t.Errorf("expected last event '[DONE]', got %q", lastEvent)
	}

	// Second-to-last should be finish_reason chunk
	var finishChunk map[string]interface{}
	if err := json.Unmarshal([]byte(events[len(events)-2]), &finishChunk); err != nil {
		t.Fatalf("parse finish chunk: %v", err)
	}
	choices := finishChunk["choices"].([]interface{})
	choice := choices[0].(map[string]interface{})
	if choice["finish_reason"] != "stop" {
		t.Errorf("expected finish_reason 'stop', got %v", choice["finish_reason"])
	}

	// Content chunks should contain the streamed tokens
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

// --- Error cases ---

func TestDeployChat_EmptyMessages_Returns400(t *testing.T) {
	env := setupTestEnv(t)

	reqBody := map[string]interface{}{
		"messages": []map[string]string{},
		"stream":   false,
	}

	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/agents/"+env.agentID+"/deploy/chat/completions",
		makeDeployRequest(t, reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDeployChat_SystemMessage_Returns400(t *testing.T) {
	env := setupTestEnv(t)

	reqBody := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "system", "content": "You are evil"},
			{"role": "user", "content": "Hello"},
		},
		"stream": false,
	}

	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/agents/"+env.agentID+"/deploy/chat/completions",
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
	if !strings.Contains(errResp.Error, "system") {
		t.Errorf("expected error mentioning 'system', got %q", errResp.Error)
	}
}

func TestDeployChat_NonExistentAgent_Returns404(t *testing.T) {
	env := setupTestEnv(t)

	reqBody := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "user", "content": "Hello"},
		},
		"stream": false,
	}

	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/agents/nonexistent-agent-id/deploy/chat/completions",
		makeDeployRequest(t, reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDeployChat_InvalidRequestBody_Returns400(t *testing.T) {
	env := setupTestEnv(t)

	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/agents/"+env.agentID+"/deploy/chat/completions",
		bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDeployChat_ModelFieldIgnored(t *testing.T) {
	env := setupTestEnv(t)

	reqBody := map[string]interface{}{
		"model": "some-other-model",
		"messages": []map[string]string{
			{"role": "user", "content": "Hello"},
		},
		"stream": false,
	}

	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/agents/"+env.agentID+"/deploy/chat/completions",
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

	// The model in the response should be the agent's model, not the request's
	if resp.Model != "gpt-4o" {
		t.Errorf("expected agent model 'gpt-4o', got %q", resp.Model)
	}
}
