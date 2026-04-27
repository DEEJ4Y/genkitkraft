package services

import (
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"

	aesgcmencryptor "github.com/DEEJ4Y/genkitkraft/internal/adapters/aesgcm_encryptor"
	bcrypthasher "github.com/DEEJ4Y/genkitkraft/internal/adapters/bcrypt_hasher"
	genkitchatprovider "github.com/DEEJ4Y/genkitkraft/internal/adapters/genkit_chat_provider"
	httpprovidertester "github.com/DEEJ4Y/genkitkraft/internal/adapters/http_provider_tester"
	mcpdiscoveryadapter "github.com/DEEJ4Y/genkitkraft/internal/adapters/mcp_discovery"
	memorysession "github.com/DEEJ4Y/genkitkraft/internal/adapters/memory_session"
	sqlitedb "github.com/DEEJ4Y/genkitkraft/internal/adapters/sqlite_db"
	sqliteagent "github.com/DEEJ4Y/genkitkraft/internal/adapters/sqlite_agent"
	sqliteagenttool "github.com/DEEJ4Y/genkitkraft/internal/adapters/sqlite_agent_tool"
	sqlitehttptool "github.com/DEEJ4Y/genkitkraft/internal/adapters/sqlite_http_tool"
	sqlitemcpserver "github.com/DEEJ4Y/genkitkraft/internal/adapters/sqlite_mcp_server"
	sqliteplayground "github.com/DEEJ4Y/genkitkraft/internal/adapters/sqlite_playground"
	sqliteprompt "github.com/DEEJ4Y/genkitkraft/internal/adapters/sqlite_prompt"
	sqliteprovider "github.com/DEEJ4Y/genkitkraft/internal/adapters/sqlite_provider"
	"github.com/DEEJ4Y/genkitkraft/internal/api/gen"
	"github.com/DEEJ4Y/genkitkraft/internal/app"
	"github.com/DEEJ4Y/genkitkraft/internal/app/commands"
	"github.com/DEEJ4Y/genkitkraft/internal/app/decorators"
	"github.com/DEEJ4Y/genkitkraft/internal/app/queries"
	"github.com/DEEJ4Y/genkitkraft/internal/config"
	domainauth "github.com/DEEJ4Y/genkitkraft/internal/domain/auth"
	httphandler "github.com/DEEJ4Y/genkitkraft/internal/handlers/http_handler"
	"github.com/DEEJ4Y/genkitkraft/internal/handlers/http_handler/interceptors"
	mcphandler "github.com/DEEJ4Y/genkitkraft/internal/handlers/mcp_handler"
	chatprovider "github.com/DEEJ4Y/genkitkraft/internal/ports/chat_provider"
	"github.com/DEEJ4Y/genkitkraft/internal/ports/hasher"
	"github.com/DEEJ4Y/genkitkraft/internal/ports/session"
)

// Server is the main HTTP server for GenKitKraft.
type Server struct {
	cfg           config.Config
	authApp       *app.AuthApp
	providerApp   *app.ProviderApp
	promptApp     *app.PromptApp
	agentApp      *app.AgentApp
	playgroundApp *app.PlaygroundApp
	httpToolApp   *app.HttpToolApp
	mcpServerApp  *app.McpServerApp
	agentToolApp  *app.AgentToolApp
	chatProvider  chatprovider.ChatProvider
	sessionStore  session.Store
	db            *sql.DB
	done          chan struct{}
}

// NewServer wires all dependencies and returns a ready-to-start Server.
func NewServer(cfg config.Config) (*Server, error) {
	if cfg.Encryption.Key == "" {
		return nil, fmt.Errorf("ENCRYPTION_KEY environment variable is required")
	}

	enc, err := aesgcmencryptor.NewAESGCMEncryptor(cfg.Encryption.Key)
	if err != nil {
		return nil, fmt.Errorf("creating encryptor: %w", err)
	}
	cfg.Encryption.Key = ""

	// Create adapters
	passwordHasher := bcrypthasher.NewBcryptHasher()
	sessionStore := memorysession.NewMemoryStore()

	// Hash credentials using adapter, then discard plaintext
	users, err := hashCredentials(cfg.Auth.Credentials, passwordHasher)
	if err != nil {
		return nil, fmt.Errorf("hashing credentials: %w", err)
	}
	cfg.Auth.Credentials = nil

	authRequired := len(users) > 0

	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()

	// Create commands
	loginCmd := commands.NewLoginCommand(users, sessionStore, passwordHasher)
	logoutCmd := commands.NewLogoutCommand(sessionStore)

	// Apply decorators
	rateLimitedLogin := decorators.NewRateLimitingLoginDecorator(loginCmd)

	// Create queries
	getMeQuery := queries.NewGetMeQuery(sessionStore)
	getAuthStatusQuery := queries.NewGetAuthStatusQuery(authRequired)

	// Build application
	authApp := &app.AuthApp{
		Commands: app.AuthCommands{
			Login:  decorators.ApplyLogging(rateLimitedLogin, "Login", logger),
			Logout: decorators.ApplyLoggingExecutor(logoutCmd, "Logout", logger),
		},
		Queries: app.AuthQueries{
			GetMe:         decorators.ApplyLogging(getMeQuery, "GetMe", logger),
			GetAuthStatus: decorators.ApplyLogging(getAuthStatusQuery, "GetAuthStatus", logger),
		},
	}

	if authRequired {
		log.Printf("Authentication enabled (%d user(s) configured)", len(users))
	} else {
		log.Printf("Authentication disabled (AUTH_CREDENTIALS not set)")
	}

	// Open database and run migrations
	db, err := sqlitedb.Open(cfg.Database.Path)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}
	log.Printf("Database opened at %s", cfg.Database.Path)

	// Create provider adapters
	providerRepo := sqliteprovider.NewProviderRepository(db)
	providerTester := httpprovidertester.NewTester()

	// Create provider commands
	createProviderCmd := commands.NewCreateProviderCommand(providerRepo, enc)
	updateProviderCmd := commands.NewUpdateProviderCommand(providerRepo, enc)
	deleteProviderCmd := commands.NewDeleteProviderCommand(providerRepo)
	testProviderCmd := commands.NewTestProviderCommand(providerRepo, providerTester, enc)

	// Create provider queries
	listProvidersQuery := queries.NewListProvidersQuery(providerRepo, enc)
	getProviderQuery := queries.NewGetProviderQuery(providerRepo, enc)
	listProviderTypesQuery := queries.NewListProviderTypesQuery()

	// Build provider application
	providerApp := &app.ProviderApp{
		Commands: app.ProviderCommands{
			CreateProvider: decorators.ApplyLogging(createProviderCmd, "CreateProvider", logger),
			UpdateProvider: decorators.ApplyLogging(updateProviderCmd, "UpdateProvider", logger),
			DeleteProvider: decorators.ApplyLoggingExecutor(deleteProviderCmd, "DeleteProvider", logger),
			TestProvider:   decorators.ApplyLogging(testProviderCmd, "TestProvider", logger),
		},
		Queries: app.ProviderQueries{
			ListProviders:     decorators.ApplyLogging(listProvidersQuery, "ListProviders", logger),
			GetProvider:       decorators.ApplyLogging(getProviderQuery, "GetProvider", logger),
			ListProviderTypes: decorators.ApplyLogging(listProviderTypesQuery, "ListProviderTypes", logger),
		},
	}

	// Create prompt adapters
	promptRepo := sqliteprompt.NewPromptRepository(db)

	// Create prompt commands
	createPromptCmd := commands.NewCreatePromptCommand(promptRepo)
	updatePromptCmd := commands.NewUpdatePromptCommand(promptRepo)
	deletePromptCmd := commands.NewDeletePromptCommand(promptRepo)

	// Create prompt queries
	listPromptsQuery := queries.NewListPromptsQuery(promptRepo)
	getPromptQuery := queries.NewGetPromptQuery(promptRepo)

	// Build prompt application
	promptApp := &app.PromptApp{
		Commands: app.PromptCommands{
			CreatePrompt: decorators.ApplyLogging(createPromptCmd, "CreatePrompt", logger),
			UpdatePrompt: decorators.ApplyLogging(updatePromptCmd, "UpdatePrompt", logger),
			DeletePrompt: decorators.ApplyLoggingExecutor(deletePromptCmd, "DeletePrompt", logger),
		},
		Queries: app.PromptQueries{
			ListPrompts: decorators.ApplyLogging(listPromptsQuery, "ListPrompts", logger),
			GetPrompt:   decorators.ApplyLogging(getPromptQuery, "GetPrompt", logger),
		},
	}

	// Create agent adapters
	agentRepo := sqliteagent.NewAgentRepository(db)

	// Create HTTP tool adapters
	httpToolRepo := sqlitehttptool.NewHttpToolRepository(db)

	// Create HTTP tool commands
	createHttpToolCmd := commands.NewCreateHttpToolCommand(httpToolRepo)
	updateHttpToolCmd := commands.NewUpdateHttpToolCommand(httpToolRepo)
	deleteHttpToolCmd := commands.NewDeleteHttpToolCommand(httpToolRepo)

	// Create HTTP tool queries
	listHttpToolsQuery := queries.NewListHttpToolsQuery(httpToolRepo)
	getHttpToolQuery := queries.NewGetHttpToolQuery(httpToolRepo)

	// Build HTTP tool application
	httpToolApp := &app.HttpToolApp{
		Commands: app.HttpToolCommands{
			CreateHttpTool: decorators.ApplyLogging(createHttpToolCmd, "CreateHttpTool", logger),
			UpdateHttpTool: decorators.ApplyLogging(updateHttpToolCmd, "UpdateHttpTool", logger),
			DeleteHttpTool: decorators.ApplyLoggingExecutor(deleteHttpToolCmd, "DeleteHttpTool", logger),
		},
		Queries: app.HttpToolQueries{
			ListHttpTools: decorators.ApplyLogging(listHttpToolsQuery, "ListHttpTools", logger),
			GetHttpTool:   decorators.ApplyLogging(getHttpToolQuery, "GetHttpTool", logger),
		},
	}

	// Create MCP server adapters
	mcpServerRepo := sqlitemcpserver.NewMcpServerRepository(db)

	// Create MCP server commands
	createMcpServerCmd := commands.NewCreateMcpServerCommand(mcpServerRepo)
	updateMcpServerCmd := commands.NewUpdateMcpServerCommand(mcpServerRepo)
	deleteMcpServerCmd := commands.NewDeleteMcpServerCommand(mcpServerRepo)

	// Create MCP server queries
	listMcpServersQuery := queries.NewListMcpServersQuery(mcpServerRepo)
	getMcpServerQuery := queries.NewGetMcpServerQuery(mcpServerRepo)

	// Build MCP server application
	mcpServerApp := &app.McpServerApp{
		Commands: app.McpServerCommands{
			CreateMcpServer: decorators.ApplyLogging(createMcpServerCmd, "CreateMcpServer", logger),
			UpdateMcpServer: decorators.ApplyLogging(updateMcpServerCmd, "UpdateMcpServer", logger),
			DeleteMcpServer: decorators.ApplyLoggingExecutor(deleteMcpServerCmd, "DeleteMcpServer", logger),
		},
		Queries: app.McpServerQueries{
			ListMcpServers: decorators.ApplyLogging(listMcpServersQuery, "ListMcpServers", logger),
			GetMcpServer:   decorators.ApplyLogging(getMcpServerQuery, "GetMcpServer", logger),
		},
	}

	// Create agent commands
	createAgentCmd := commands.NewCreateAgentCommand(agentRepo, providerRepo, promptRepo)
	updateAgentCmd := commands.NewUpdateAgentCommand(agentRepo, providerRepo, promptRepo)
	deleteAgentCmd := commands.NewDeleteAgentCommand(agentRepo)

	// Create agent queries
	listAgentsQuery := queries.NewListAgentsQuery(agentRepo)
	getAgentQuery := queries.NewGetAgentQuery(agentRepo)

	// Build agent application
	agentApp := &app.AgentApp{
		Commands: app.AgentCommands{
			CreateAgent: decorators.ApplyLogging(createAgentCmd, "CreateAgent", logger),
			UpdateAgent: decorators.ApplyLogging(updateAgentCmd, "UpdateAgent", logger),
			DeleteAgent: decorators.ApplyLoggingExecutor(deleteAgentCmd, "DeleteAgent", logger),
		},
		Queries: app.AgentQueries{
			ListAgents: decorators.ApplyLogging(listAgentsQuery, "ListAgents", logger),
			GetAgent:   decorators.ApplyLogging(getAgentQuery, "GetAgent", logger),
		},
	}

	// Create agent tool adapters
	agentToolRepo := sqliteagenttool.NewRepository(db)

	// Create agent tool commands and queries
	updateAgentToolsCmd := commands.NewUpdateAgentToolsCommand(agentToolRepo)
	getAgentToolsQuery := queries.NewGetAgentToolsQuery(agentToolRepo)

	// Build agent tool application
	agentToolApp := &app.AgentToolApp{
		Commands: app.AgentToolCommands{
			UpdateTools: decorators.ApplyLogging(updateAgentToolsCmd, "UpdateAgentTools", logger),
		},
		Queries: app.AgentToolQueries{
			GetTools: decorators.ApplyLogging(getAgentToolsQuery, "GetAgentTools", logger),
		},
	}

	// Create playground adapters
	playgroundRepo := sqliteplayground.NewPlaygroundRepository(db)
	chatProvider := genkitchatprovider.NewChatProvider()

	// Create playground commands
	createSessionCmd := commands.NewCreatePlaygroundSessionCommand(playgroundRepo, agentRepo)
	deleteSessionCmd := commands.NewDeletePlaygroundSessionCommand(playgroundRepo)
	saveMessageCmd := commands.NewSavePlaygroundMessageCommand(playgroundRepo)

	// Create playground queries
	listSessionsQuery := queries.NewListPlaygroundSessionsQuery(playgroundRepo)
	getSessionQuery := queries.NewGetPlaygroundSessionQuery(playgroundRepo)
	listMessagesQuery := queries.NewListPlaygroundMessagesQuery(playgroundRepo)
	resolveConfigQuery := queries.NewResolvePlaygroundConfigQuery(agentRepo, providerRepo, promptRepo, enc, agentToolRepo, httpToolRepo, mcpServerRepo)

	// Build playground application
	playgroundApp := &app.PlaygroundApp{
		Commands: app.PlaygroundCommands{
			CreateSession: decorators.ApplyLogging(createSessionCmd, "CreatePlaygroundSession", logger),
			DeleteSession: decorators.ApplyLoggingExecutor(deleteSessionCmd, "DeletePlaygroundSession", logger),
			SaveMessage:   decorators.ApplyLogging(saveMessageCmd, "SavePlaygroundMessage", logger),
		},
		Queries: app.PlaygroundQueries{
			ListSessions:  decorators.ApplyLogging(listSessionsQuery, "ListPlaygroundSessions", logger),
			GetSession:    decorators.ApplyLogging(getSessionQuery, "GetPlaygroundSession", logger),
			ListMessages:  decorators.ApplyLogging(listMessagesQuery, "ListPlaygroundMessages", logger),
			ResolveConfig: decorators.ApplyLogging(resolveConfigQuery, "ResolvePlaygroundConfig", logger),
		},
	}

	return &Server{
		cfg:           cfg,
		authApp:       authApp,
		providerApp:   providerApp,
		promptApp:     promptApp,
		agentApp:      agentApp,
		playgroundApp: playgroundApp,
		httpToolApp:   httpToolApp,
		mcpServerApp:  mcpServerApp,
		agentToolApp:  agentToolApp,
		chatProvider:  chatProvider,
		sessionStore:  sessionStore,
		db:            db,
		done:          make(chan struct{}),
	}, nil
}

// Start begins serving HTTP. This blocks until the server stops.
func (s *Server) Start() error {
	s.sessionStore.StartCleanupLoop(s.done)

	mux := http.NewServeMux()

	// Register all API routes via generated handler
	mcpDiscovery := mcpdiscoveryadapter.New()
	apiHandler := httphandler.NewHandler(s.authApp, s.providerApp, s.promptApp, s.agentApp, s.playgroundApp, s.httpToolApp, s.mcpServerApp, s.agentToolApp, s.chatProvider, mcpDiscovery)
	gen.HandlerFromMux(apiHandler, mux)

	// SPA fallback: serve embedded UI or fallback to index.html
	mux.HandleFunc("/", spaHandler())

	// Mount MCP server at /mcp
	mcpH := mcphandler.NewHandler(
		s.authApp, s.providerApp, s.promptApp, s.agentApp,
		s.playgroundApp, s.httpToolApp, s.mcpServerApp, s.agentToolApp,
		s.chatProvider, mcpDiscovery, s.cfg.Auth,
	)
	mux.Handle("/mcp", mcpH.HTTPHandler())

	// Wrap with auth middleware
	handler := interceptors.DeployAuthMiddleware(s.cfg.Deploy.APIKeys)(interceptors.AuthMiddleware(s.authApp)(mux))

	addr := ":" + s.cfg.Server.Port
	log.Printf("Server listening on %s", addr)
	return http.ListenAndServe(addr, handler)
}

// Stop signals background goroutines to stop and releases resources.
func (s *Server) Stop() {
	close(s.done)
	if s.db != nil {
		s.db.Close()
	}
}

// hashCredentials takes config credentials and returns a user map for the login command.
func hashCredentials(creds []config.AuthCredential, h hasher.PasswordHasher) (map[string]*domainauth.User, error) {
	users := make(map[string]*domainauth.User, len(creds))
	for _, c := range creds {
		hash, err := h.Hash(c.Password)
		if err != nil {
			return nil, fmt.Errorf("hashing password for %s: %w", c.Username, err)
		}
		users[c.Username] = &domainauth.User{
			Username:     c.Username,
			PasswordHash: hash,
		}
	}
	return users, nil
}

// spaHandler serves static files from ui/dist if it exists,
// with fallback to index.html for client-side routing.
func spaHandler() http.HandlerFunc {
	distPath := "ui/dist"

	if _, err := os.Stat(distPath); os.IsNotExist(err) {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message":"GenKitKraft API is running. UI not built yet."}`))
		}
	}

	staticFS := os.DirFS(distPath)

	serveFile := func(w http.ResponseWriter, fsys fs.FS, name string) {
		data, err := fs.ReadFile(fsys, name)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		if f, err := fs.Stat(staticFS, path); err == nil && !f.IsDir() {
			http.ServeFileFS(w, r, staticFS, path)
			return
		}

		// Next.js static export generates {page}.html files for each route.
		// Try serving the .html version before falling back to index.html.
		if filepath.Ext(path) == "" {
			htmlPath := path + ".html"
			if _, err := fs.Stat(staticFS, htmlPath); err == nil {
				serveFile(w, staticFS, htmlPath)
				return
			}
		}

		if filepath.Ext(path) != "" {
			http.NotFound(w, r)
			return
		}

		serveFile(w, staticFS, "index.html")
	}
}
