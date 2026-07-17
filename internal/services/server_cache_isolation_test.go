//go:build integration

package services

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/DEEJ4Y/genkitkraft/internal/app/commands"
	"github.com/DEEJ4Y/genkitkraft/internal/config"
	"github.com/DEEJ4Y/genkitkraft/resources/test/containers"
)

// newSharedCacheServer boots a full server against real Postgres and Valkey — the
// multi-instance configuration this whole feature exists to support.
func newSharedCacheServer(t *testing.T) (*Server, *redis.Client) {
	t.Helper()

	cacheURL := containers.StartValkeyURL(t)
	cfg := config.Config{
		Server: config.ServerConfig{Port: "0"},
		Auth: config.AuthConfig{
			Credentials: []config.AuthCredential{{Username: "alice", Password: "secret"}},
		},
		Database: config.DatabaseConfig{
			Provider: "postgres",
			URL:      containers.StartPostgresDSN(t),
		},
		Cache:      config.CacheConfig{Provider: "valkey", URL: cacheURL},
		Encryption: config.EncryptionConfig{Key: "test-encryption-key"},
	}

	srv, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	t.Cleanup(srv.Stop)

	opts, err := redis.ParseURL(cacheURL)
	if err != nil {
		t.Fatalf("ParseURL(%q): %v", cacheURL, err)
	}
	client := redis.NewClient(opts)
	t.Cleanup(func() { client.Close() })

	return srv, client
}

func keysWithPrefix(t *testing.T, client *redis.Client, pattern string) []string {
	t.Helper()
	keys, err := client.Keys(context.Background(), pattern).Result()
	if err != nil {
		t.Fatalf("KEYS %s: %v", pattern, err)
	}
	return keys
}

// The built-in web_fetch tool must never write to the shared cache. Its keys are
// URLs chosen by the model and its values are whole fetched pages, so if it shared
// the store that holds sessions, agent traffic could evict session tokens or fill
// the instance — and under the recommended noeviction policy that stops every
// login. Keeping the two apart is the entire point of the separate store, so it is
// asserted here rather than left to a comment somebody later "simplifies" away.
func TestWebFetchCacheIsNotTheSharedStore(t *testing.T) {
	srv, client := newSharedCacheServer(t)
	ctx := context.Background()

	// Positive control first. Without proving that the shared store really is wired
	// up and reachable, the web_fetch assertion below would pass just as happily
	// against a Valkey nobody ever talks to.
	if _, err := srv.authApp.Commands.Login.Execute(ctx, commands.LoginParams{
		Username: "alice",
		Password: "secret",
		ClientIP: "203.0.113.7",
	}); err != nil {
		t.Fatalf("login: %v", err)
	}
	if got := keysWithPrefix(t, client, "session:*"); len(got) == 0 {
		t.Fatal("no session:* key in the shared cache after a successful login; the shared store is not actually wired up, so this test could not detect a web_fetch leak either")
	}

	// The real assertion: a write through the web_fetch cache must not reach Valkey.
	if err := srv.webFetchStore.Scope("web_fetch").Set(ctx, "https://example.com/page", "# a fetched page", time.Hour); err != nil {
		t.Fatalf("write to the web_fetch cache: %v", err)
	}
	if got := keysWithPrefix(t, client, "web_fetch:*"); len(got) != 0 {
		t.Errorf("web_fetch wrote %v into the shared cache; it must stay process-local so agent traffic cannot evict sessions", got)
	}
}

// A weaker but sharper guard on the same regression: even when CACHE_PROVIDER is
// memory — where both stores are the same type and sharing one would look harmless
// — web_fetch must still get its own. Otherwise the isolation would hold only in
// some configurations, which is how the original bug got in.
func TestWebFetchStoreIsDistinctFromTheSharedStore(t *testing.T) {
	srv, _ := newSharedCacheServer(t)

	if srv.webFetchStore == srv.cacheStore {
		t.Error("web_fetch shares the store that holds sessions and rate-limit counters")
	}
}
