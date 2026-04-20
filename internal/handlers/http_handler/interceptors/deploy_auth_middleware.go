package interceptors

import (
	"encoding/json"
	"net/http"
	"strings"
)

// isDeployPath returns true if the request path is a deploy API endpoint.
func isDeployPath(path string) bool {
	// Deploy paths follow the pattern: /api/v1/agents/{agentId}/deploy/...
	const prefix = "/api/v1/agents/"
	const segment = "/deploy/"
	if !strings.HasPrefix(path, prefix) {
		return false
	}
	rest := path[len(prefix):]
	return strings.Contains(rest, segment)
}

type openAIError struct {
	Error openAIErrorBody `json:"error"`
}

type openAIErrorBody struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// DeployAuthMiddleware returns HTTP middleware that enforces API key authentication
// on deploy endpoints. If apiKeys is empty, all deploy endpoints are publicly accessible.
func DeployAuthMiddleware(apiKeys map[string]struct{}) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !isDeployPath(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			if len(apiKeys) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			auth := r.Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Bearer ") {
				writeDeployUnauthorized(w)
				return
			}

			key := strings.TrimPrefix(auth, "Bearer ")
			if _, ok := apiKeys[key]; !ok {
				writeDeployUnauthorized(w)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func writeDeployUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(openAIError{
		Error: openAIErrorBody{
			Message: "Invalid API key",
			Type:    "invalid_request_error",
			Code:    "invalid_api_key",
		},
	})
}
