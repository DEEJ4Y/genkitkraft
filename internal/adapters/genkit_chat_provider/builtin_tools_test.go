package genkitchatprovider

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// The URL is chosen by the model, so the response size is attacker-influenced: the
// body must be capped rather than read whole into memory.
func TestFetchAsMarkdown_CapsLargeBody(t *testing.T) {
	body := "<html><body><p>" + strings.Repeat("a", 4<<20) + "</p></body></html>"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(body))
	}))
	defer server.Close()

	got, err := fetchAsMarkdown(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) > maxResponseBytes+200 {
		t.Errorf("markdown is %d bytes from a %d-byte body; want at most ~%d — the read was not capped",
			len(got), len(body), maxResponseBytes)
	}
	if !strings.Contains(got, "[Content truncated") {
		t.Error("a truncated fetch must say so; a model summarizing this would present a partial page as complete")
	}
}

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
