package interceptors_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DEEJ4Y/genkitkraft/internal/handlers/http_handler/interceptors"
)

func TestDeployAuthMiddleware_NoKeys_PublicAccess(t *testing.T) {
	handler := interceptors.DeployAuthMiddleware(map[string]struct{}{})(okHandler())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents/abc/deploy/chat/completions", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestDeployAuthMiddleware_ValidBearer(t *testing.T) {
	keys := map[string]struct{}{"test-key-1": {}, "test-key-2": {}}
	handler := interceptors.DeployAuthMiddleware(keys)(okHandler())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents/abc/deploy/chat/completions", nil)
	req.Header.Set("Authorization", "Bearer test-key-1")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestDeployAuthMiddleware_InvalidBearer(t *testing.T) {
	keys := map[string]struct{}{"valid-key": {}}
	handler := interceptors.DeployAuthMiddleware(keys)(okHandler())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents/abc/deploy/chat/completions", nil)
	req.Header.Set("Authorization", "Bearer wrong-key")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}

	var body map[string]map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	if body["error"]["code"] != "invalid_api_key" {
		t.Fatalf("expected error code 'invalid_api_key', got %q", body["error"]["code"])
	}
}

func TestDeployAuthMiddleware_MissingAuthHeader(t *testing.T) {
	keys := map[string]struct{}{"valid-key": {}}
	handler := interceptors.DeployAuthMiddleware(keys)(okHandler())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents/abc/deploy/chat/completions", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestDeployAuthMiddleware_NonDeployPath_PassesThrough(t *testing.T) {
	keys := map[string]struct{}{"valid-key": {}}
	handler := interceptors.DeployAuthMiddleware(keys)(okHandler())

	// Non-deploy path should pass through without auth
	req := httptest.NewRequest(http.MethodGet, "/api/v1/agents", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for non-deploy path, got %d", w.Code)
	}
}

func TestDeployAuthMiddleware_HistoryEndpoints_RequireKey(t *testing.T) {
	keys := map[string]struct{}{"valid-key": {}}
	handler := interceptors.DeployAuthMiddleware(keys)(okHandler())

	paths := []string{
		"/api/v1/agents/abc/deploy/sessions",
		"/api/v1/agents/abc/deploy/sessions?limit=5&offset=10",
		"/api/v1/agents/abc/deploy/sessions/s1/messages",
	}
	for _, path := range paths {
		for name, header := range map[string]string{"missing": "", "wrong": "Bearer nope", "valid": "Bearer valid-key"} {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			if header != "" {
				req.Header.Set("Authorization", header)
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			want := http.StatusUnauthorized
			if name == "valid" {
				want = http.StatusOK
			}
			if w.Code != want {
				t.Errorf("%s with %s key: expected %d, got %d", path, name, want, w.Code)
			}
		}
	}
}

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}
