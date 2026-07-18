package genkitchatprovider

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// stubCache lets each web_fetch test dictate what the cache lookup reports.
type stubCache struct {
	getVal string
	getOK  bool
	getErr error
}

func (s *stubCache) Get(context.Context, string) (string, bool, error) {
	return s.getVal, s.getOK, s.getErr
}
func (s *stubCache) Set(context.Context, string, string, time.Duration) error { return nil }
func (s *stubCache) Delete(context.Context, string) error                     { return nil }
func (s *stubCache) Increment(context.Context, string, time.Duration) (int64, error) {
	return 0, nil
}
func (s *stubCache) Decrement(context.Context, string) error { return nil }

// Unlike session lookup, a cache error here must not fail the call: web_fetch is a
// pure cache, so an outage should cost a round trip rather than break the tool.
func TestWebFetch_CacheErrorRefetches(t *testing.T) {
	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		w.Write([]byte("<html><body><h1>Hello</h1></body></html>"))
	}))
	defer server.Close()

	cp := &ChatProvider{cache: &stubCache{getErr: io.ErrUnexpectedEOF}}

	got, err := cp.webFetch(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("webFetch with a failing cache: %v", err)
	}
	if !strings.Contains(got, "Hello") {
		t.Errorf("webFetch = %q, want the fetched page", got)
	}
	if n := hits.Load(); n != 1 {
		t.Errorf("origin was hit %d times, want 1 — a cache error must fall through to a fetch", n)
	}
}

// The other half: giving Get an error return must not accidentally disable caching.
func TestWebFetch_CacheHitSkipsFetch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("origin was fetched despite a cache hit")
	}))
	defer server.Close()

	cp := &ChatProvider{cache: &stubCache{getVal: "cached markdown", getOK: true}}

	got, err := cp.webFetch(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("webFetch: %v", err)
	}
	if got != "cached markdown" {
		t.Errorf("webFetch = %q, want the cached value", got)
	}
}

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
