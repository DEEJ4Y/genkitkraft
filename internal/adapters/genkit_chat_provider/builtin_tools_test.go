package genkitchatprovider

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchAsMarkdown_Non2xxStatus(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       string
	}{
		{"404 Not Found", http.StatusNotFound, "Fetch failed with status 404"},
		{"500 Internal Server Error", http.StatusInternalServerError, "Fetch failed with status 500"},
		{"400 Bad Request", http.StatusBadRequest, "Fetch failed with status 400"},
		{"403 Forbidden", http.StatusForbidden, "Fetch failed with status 403"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte("<html><body>Error page</body></html>"))
			}))
			defer server.Close()

			got, err := fetchAsMarkdown(server.URL)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFetchAsMarkdown_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body><h1>Hello</h1><p>World</p></body></html>"))
	}))
	defer server.Close()

	got, err := fetchAsMarkdown(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == "" {
		t.Error("expected non-empty markdown output")
	}
}

func TestFetchAsMarkdown_EmptyContent(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"empty body", ""},
		{"empty html", "<html><body></body></html>"},
		{"whitespace only", "   \n\t  "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tt.body))
			}))
			defer server.Close()

			got, err := fetchAsMarkdown(server.URL)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			want := "No content received. Page may be dynamically rendered."
			if got != want {
				t.Errorf("got %q, want %q", got, want)
			}
		})
	}
}
