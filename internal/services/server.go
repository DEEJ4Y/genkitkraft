package services

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	bcrypthasher "github.com/DEEJ4Y/genkitkraft/internal/adapters/bcrypt_hasher"
	memorysession "github.com/DEEJ4Y/genkitkraft/internal/adapters/memory_session"
	"github.com/DEEJ4Y/genkitkraft/internal/api/gen"
	"github.com/DEEJ4Y/genkitkraft/internal/app"
	"github.com/DEEJ4Y/genkitkraft/internal/app/commands"
	"github.com/DEEJ4Y/genkitkraft/internal/app/decorators"
	"github.com/DEEJ4Y/genkitkraft/internal/app/queries"
	"github.com/DEEJ4Y/genkitkraft/internal/config"
	domainauth "github.com/DEEJ4Y/genkitkraft/internal/domain/auth"
	httphandler "github.com/DEEJ4Y/genkitkraft/internal/handlers/http_handler"
	"github.com/DEEJ4Y/genkitkraft/internal/handlers/http_handler/interceptors"
	"github.com/DEEJ4Y/genkitkraft/internal/ports/hasher"
	"github.com/DEEJ4Y/genkitkraft/internal/ports/session"
)

// Server is the main HTTP server for GenKitKraft.
type Server struct {
	cfg          config.Config
	authApp      *app.AuthApp
	sessionStore session.Store
	done         chan struct{}
}

// NewServer wires all dependencies and returns a ready-to-start Server.
func NewServer(cfg config.Config) (*Server, error) {
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

	// Create commands
	loginCmd := commands.NewLoginCommand(users, sessionStore, passwordHasher)
	logoutCmd := commands.NewLogoutCommand(sessionStore)

	// Apply decorators
	decoratedLogin := decorators.NewRateLimitingLoginDecorator(loginCmd)

	// Create queries
	getMeQuery := queries.NewGetMeQuery(sessionStore)
	getAuthStatusQuery := queries.NewGetAuthStatusQuery(authRequired)

	// Build application
	authApp := &app.AuthApp{
		Commands: app.AuthCommands{
			Login:  decoratedLogin,
			Logout: logoutCmd,
		},
		Queries: app.AuthQueries{
			GetMe:         getMeQuery,
			GetAuthStatus: getAuthStatusQuery,
		},
	}

	if authRequired {
		log.Printf("Authentication enabled (%d user(s) configured)", len(users))
	} else {
		log.Printf("Authentication disabled (AUTH_CREDENTIALS not set)")
	}

	return &Server{
		cfg:          cfg,
		authApp:      authApp,
		sessionStore: sessionStore,
		done:         make(chan struct{}),
	}, nil
}

// Start begins serving HTTP. This blocks until the server stops.
func (s *Server) Start() error {
	s.sessionStore.StartCleanupLoop(s.done)

	mux := http.NewServeMux()

	// Register all API routes via generated handler
	apiHandler := httphandler.NewHandler(s.authApp)
	gen.HandlerFromMux(apiHandler, mux)

	// SPA fallback: serve embedded UI or fallback to index.html
	mux.HandleFunc("/", spaHandler())

	// Wrap with auth middleware
	handler := interceptors.AuthMiddleware(s.authApp)(mux)

	addr := ":" + s.cfg.Server.Port
	log.Printf("Server listening on %s", addr)
	return http.ListenAndServe(addr, handler)
}

// Stop signals background goroutines to stop.
func (s *Server) Stop() {
	close(s.done)
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

	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		if f, err := fs.Stat(staticFS, path); err == nil && !f.IsDir() {
			http.ServeFileFS(w, r, staticFS, path)
			return
		}

		if filepath.Ext(path) != "" {
			http.NotFound(w, r)
			return
		}

		http.ServeFileFS(w, r, staticFS, "index.html")
	}
}
