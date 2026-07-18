package interceptors

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/DEEJ4Y/genkitkraft/internal/api/gen"
	"github.com/DEEJ4Y/genkitkraft/internal/app"
	"github.com/DEEJ4Y/genkitkraft/internal/app/queries"
	"github.com/DEEJ4Y/genkitkraft/internal/common/errors"
)

type contextKey string

const (
	usernameContextKey contextKey = "auth_username"
	sessionCookieName             = "session_token"
)

// UsernameFromContext extracts the authenticated username from the request context.
func UsernameFromContext(ctx context.Context) (string, bool) {
	username, ok := ctx.Value(usernameContextKey).(string)
	return username, ok
}

// AuthMiddleware returns HTTP middleware that enforces authentication on protected routes.
func AuthMiddleware(authApp *app.AuthApp) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			statusResult, _ := authApp.Queries.GetAuthStatus.Execute(r.Context(), queries.GetAuthStatusParams{})
			if !statusResult.Required {
				next.ServeHTTP(w, r)
				return
			}

			if isPublicPath(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			if strings.HasPrefix(r.URL.Path, "/v1/") {
				next.ServeHTTP(w, r)
				return
			}

			// Deploy endpoints use their own API key auth via DeployAuthMiddleware
			if isDeployPath(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			if !strings.HasPrefix(r.URL.Path, "/api/") {
				next.ServeHTTP(w, r)
				return
			}

			cookie, err := r.Cookie(sessionCookieName)
			if err != nil {
				writeUnauthorized(w)
				return
			}

			result, err := authApp.Queries.GetMe.Execute(r.Context(), queries.GetMeParams{Token: cookie.Value})
			if err != nil {
				writeAuthError(w, err)
				return
			}

			ctx := context.WithValue(r.Context(), usernameContextKey, result.Username)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func isPublicPath(path string) bool {
	switch path {
	case "/api/auth/status", "/api/auth/login":
		return true
	}
	return false
}

// writeAuthError surfaces a dependency outage and fails closed on everything else.
//
// A cache outage is not a bad credential: answering 401 sends the user to a login
// that returns 503 anyway, so the two halves of the API would describe different
// failures. Every other code deliberately collapses to 401 rather than going
// through writeAppError — this is the auth gate, and a session lookup has no
// business returning a 404 or a 409 to an unauthenticated caller.
func writeAuthError(w http.ResponseWriter, err error) {
	if appErr, ok := errors.IsAppError(err); ok && appErr.Code() == errors.Unavailable {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(errors.HTTPStatusCode(appErr.Code()))
		json.NewEncoder(w).Encode(gen.ModelsErrorResponse{Error: appErr.Error()})
		return
	}
	writeUnauthorized(w)
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(gen.ModelsErrorResponse{Error: "unauthorized"})
}
