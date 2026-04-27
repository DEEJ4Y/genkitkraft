package mcphandler

import (
	"crypto/subtle"
	"net/http"

	"github.com/DEEJ4Y/genkitkraft/internal/config"
)

// basicAuthMiddleware wraps an http.Handler with HTTP Basic Authentication.
// It validates the provided username:password against the configured credentials.
func basicAuthMiddleware(next http.Handler, credentials []config.AuthCredential) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok {
			w.Header().Set("WWW-Authenticate", `Basic realm="GenKitKraft MCP"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if !validateCredentials(username, password, credentials) {
			w.Header().Set("WWW-Authenticate", `Basic realm="GenKitKraft MCP"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func validateCredentials(username, password string, credentials []config.AuthCredential) bool {
	for _, cred := range credentials {
		if subtle.ConstantTimeCompare([]byte(username), []byte(cred.Username)) == 1 &&
			subtle.ConstantTimeCompare([]byte(password), []byte(cred.Password)) == 1 {
			return true
		}
	}
	return false
}
