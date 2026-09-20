package mcphandler

import (
	"context"
	"path/filepath"
	"testing"

	aesgcmencryptor "github.com/DEEJ4Y/genkitkraft/internal/adapters/aesgcm_encryptor"
	sqliteagent "github.com/DEEJ4Y/genkitkraft/internal/adapters/sqlite_agent"
	sqliteagenttool "github.com/DEEJ4Y/genkitkraft/internal/adapters/sqlite_agent_tool"
	sqlitedb "github.com/DEEJ4Y/genkitkraft/internal/adapters/sqlite_db"
	sqlitehttptool "github.com/DEEJ4Y/genkitkraft/internal/adapters/sqlite_http_tool"
	sqlitemcpserver "github.com/DEEJ4Y/genkitkraft/internal/adapters/sqlite_mcp_server"
	sqliteplayground "github.com/DEEJ4Y/genkitkraft/internal/adapters/sqlite_playground"
	sqliteprompt "github.com/DEEJ4Y/genkitkraft/internal/adapters/sqlite_prompt"
	sqliteprovider "github.com/DEEJ4Y/genkitkraft/internal/adapters/sqlite_provider"
	"github.com/DEEJ4Y/genkitkraft/internal/app"
	"github.com/DEEJ4Y/genkitkraft/internal/app/commands"
	"github.com/DEEJ4Y/genkitkraft/internal/app/queries"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/agent"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/prompt"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/provider"
	agentrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/agent_repo"
	"github.com/DEEJ4Y/genkitkraft/resources/test/mock"
)

// playgroundToolTestEnv wires a *Handler with a real, SQLite-backed
// PlaygroundApp (so ResolvePlaygroundConfigQuery resolves a real agent) and a
// mock ChatProvider, so tests can assert on the ChatRequest playgroundChat
// actually sends.
type playgroundToolTestEnv struct {
	handler   *Handler
	mockChat  *mock.ChatProvider
	agentID   string
	sessionID string
	agentRepo agentrepo.AgentRepository
}

func setupPlaygroundToolTestEnv(t *testing.T) *playgroundToolTestEnv {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := sqlitedb.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

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

	pr := &prompt.Prompt{Name: "test-prompt", Content: "You are a helpful assistant."}
	if err := promptRepo.Create(ctx, pr); err != nil {
		t.Fatalf("create prompt: %v", err)
	}

	a := &agent.Agent{
		Name:           "test-agent",
		ProviderID:     p.ID,
		ModelID:        "gpt-4o",
		SystemPromptID: pr.ID,
	}
	if err := agentRepo.Create(ctx, a); err != nil {
		t.Fatalf("create agent: %v", err)
	}

	mockCP := &mock.ChatProvider{ChatResponse: "Hello from mock!"}

	resolveConfig := queries.NewResolvePlaygroundConfigQuery(agentRepo, providerRepo, promptRepo, enc, agentToolRepo, httpToolRepo, mcpServerRepo)
	saveMessage := commands.NewSavePlaygroundMessageCommand(playgroundRepo)
	listMessages := queries.NewListPlaygroundMessagesQuery(playgroundRepo)

	playgroundApp := &app.PlaygroundApp{
		Commands: app.PlaygroundCommands{
			SaveMessage: saveMessage,
		},
		Queries: app.PlaygroundQueries{
			ListMessages:  listMessages,
			ResolveConfig: resolveConfig,
		},
	}

	handler := &Handler{
		playgroundApp: playgroundApp,
		chatProvider:  mockCP,
	}

	createSession := commands.NewCreatePlaygroundSessionCommand(playgroundRepo, agentRepo)
	sessionResult, err := createSession.Execute(ctx, commands.CreatePlaygroundSessionParams{AgentID: a.ID})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	return &playgroundToolTestEnv{
		handler:   handler,
		mockChat:  mockCP,
		agentID:   a.ID,
		sessionID: sessionResult.Session.ID,
		agentRepo: agentRepo,
	}
}

// Regression test: the MCP playground_chat tool used to build a ChatRequest
// without ever setting SessionID, so report_gap references from this path
// (internal/adapters/genkit_chat_provider/builtin_tools.go) could never be
// attached to a conversation, no matter the agent's gap_reporting_enabled
// flag or its assigned tools.
func TestPlaygroundChatTool_SetsSessionID(t *testing.T) {
	env := setupPlaygroundToolTestEnv(t)

	_, _, err := env.handler.playgroundChat(context.Background(), nil, PlaygroundChatInput{
		AgentID:   env.agentID,
		SessionID: env.sessionID,
		Content:   "hi",
	})
	if err != nil {
		t.Fatalf("playgroundChat: %v", err)
	}
	if env.mockChat.LastRequest.SessionID != env.sessionID {
		t.Errorf("LastRequest.SessionID = %q, want %q", env.mockChat.LastRequest.SessionID, env.sessionID)
	}
}

func TestPlaygroundChatTool_GapReportingEnabled_ReachesProvider(t *testing.T) {
	env := setupPlaygroundToolTestEnv(t)

	a, err := env.agentRepo.GetByID(context.Background(), env.agentID)
	if err != nil {
		t.Fatalf("get agent: %v", err)
	}
	a.GapReportingEnabled = true
	if err := env.agentRepo.Update(context.Background(), a); err != nil {
		t.Fatalf("enable gap reporting: %v", err)
	}

	_, _, err = env.handler.playgroundChat(context.Background(), nil, PlaygroundChatInput{
		AgentID:   env.agentID,
		SessionID: env.sessionID,
		Content:   "hi",
	})
	if err != nil {
		t.Fatalf("playgroundChat: %v", err)
	}
	if !env.mockChat.LastRequest.GapReportingEnabled {
		t.Errorf("LastRequest.GapReportingEnabled = false, want true")
	}
	if env.mockChat.LastRequest.SessionID != env.sessionID {
		t.Errorf("LastRequest.SessionID = %q, want %q", env.mockChat.LastRequest.SessionID, env.sessionID)
	}
	if env.mockChat.LastRequest.AgentID != env.agentID {
		t.Errorf("LastRequest.AgentID = %q, want %q", env.mockChat.LastRequest.AgentID, env.agentID)
	}
}
