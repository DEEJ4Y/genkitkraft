//go:build integration

package rediscache_test

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	rediscache "github.com/DEEJ4Y/genkitkraft/internal/adapters/redis_cache"
	"github.com/DEEJ4Y/genkitkraft/internal/ports/cache"
	"github.com/DEEJ4Y/genkitkraft/resources/test/cachecontract"
	"github.com/DEEJ4Y/genkitkraft/resources/test/containers"
)

func TestRedisCache(t *testing.T) {
	testCache(t, containers.StartRedisURL(t))
}

func TestValkeyCache(t *testing.T) {
	testCache(t, containers.StartValkeyURL(t))
}

func testCache(t *testing.T, url string) {
	t.Helper()
	cachecontract.Run(t, func(t *testing.T) cache.Store {
		// Subtests share one container, so clear the keyspace to give each the
		// empty store the contract expects.
		flushDB(t, url)

		store, err := rediscache.NewCache(url, zerolog.New(io.Discard))
		if err != nil {
			t.Fatalf("NewCache: %v", err)
		}
		t.Cleanup(func() { store.Close() })
		return store
	})
}

// The Redis heal happens inside the Lua script, and a heal and an ordinary first
// create both return 1 — the count alone cannot tell them apart. These two tests
// are what prove the script's second return value actually discriminates; without
// the negative case, returning a constant 1 for healed would pass.
func TestIncrementLogsWhenHealingANonCounter(t *testing.T) {
	url := containers.StartRedisURL(t)
	flushDB(t, url)

	var buf bytes.Buffer
	store, err := rediscache.NewCache(url, zerolog.New(&buf))
	if err != nil {
		t.Fatalf("NewCache: %v", err)
	}
	defer store.Close()

	c := store.Scope("rate_limit")
	ctx := context.Background()
	if err := c.Set(ctx, "1.2.3.4", "not-a-number", time.Minute); err != nil {
		t.Fatalf("Set: %v", err)
	}
	n, err := c.Increment(ctx, "1.2.3.4", time.Minute)
	if err != nil {
		t.Fatalf("Increment: %v", err)
	}
	if n != 1 {
		t.Errorf("Increment over a non-counter = %d, want 1 (a fresh counter)", n)
	}

	logged := buf.String()
	if !strings.Contains(logged, "non-counter") {
		t.Errorf("heal was not logged; log = %q", logged)
	}
	if !strings.Contains(logged, "rate_limit:1.2.3.4") {
		t.Errorf("heal log does not name the fully-qualified key; log = %q", logged)
	}
}

func TestIncrementDoesNotLogWhenCreatingACounter(t *testing.T) {
	url := containers.StartRedisURL(t)
	flushDB(t, url)

	var buf bytes.Buffer
	store, err := rediscache.NewCache(url, zerolog.New(&buf))
	if err != nil {
		t.Fatalf("NewCache: %v", err)
	}
	defer store.Close()

	if _, err := store.Scope("rate_limit").Increment(context.Background(), "1.2.3.4", time.Minute); err != nil {
		t.Fatalf("Increment: %v", err)
	}

	if buf.Len() != 0 {
		t.Errorf("creating a counter logged %q, want nothing", buf.String())
	}
}

// Valkey deployments are commonly addressed with a valkey:// URL, which go-redis
// does not recognise — the adapter must accept it.
func TestValkeySchemeIsAccepted(t *testing.T) {
	url := containers.StartValkeyURL(t)
	valkeyURL := "valkey://" + strings.TrimPrefix(url, "redis://")

	store, err := rediscache.NewCache(valkeyURL, zerolog.New(io.Discard))
	if err != nil {
		t.Fatalf("NewCache(%q): %v", valkeyURL, err)
	}
	defer store.Close()

	c := store.Scope("ns")
	ctx := context.Background()
	if err := c.Set(ctx, "key", "value", time.Minute); err != nil {
		t.Fatalf("Set: %v", err)
	}
	got, ok, err := c.Get(ctx, "key")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !ok || got != "value" {
		t.Errorf("Get = (%q, %v), want (%q, true)", got, ok, "value")
	}
}

// A bad URL must fail at construction, so a misconfigured deployment dies at boot
// rather than serving 401s once traffic arrives.
func TestNewCacheFailsOnUnreachableServer(t *testing.T) {
	// Port 1 is reserved and never listening.
	if _, err := rediscache.NewCache("redis://127.0.0.1:1", zerolog.New(io.Discard)); err == nil {
		t.Error("NewCache against an unreachable server: want error, got nil")
	}
}

func TestNewCacheFailsOnMalformedURL(t *testing.T) {
	if _, err := rediscache.NewCache("not-a-url", zerolog.New(io.Discard)); err == nil {
		t.Error("NewCache with a malformed URL: want error, got nil")
	}
}

// Two Cache instances against one server model two application instances; the
// point of the Redis adapter is that they see each other's writes.
func TestSeparateInstancesShareState(t *testing.T) {
	url := containers.StartRedisURL(t)
	flushDB(t, url)

	instanceA, err := rediscache.NewCache(url, zerolog.New(io.Discard))
	if err != nil {
		t.Fatalf("NewCache: %v", err)
	}
	defer instanceA.Close()

	instanceB, err := rediscache.NewCache(url, zerolog.New(io.Discard))
	if err != nil {
		t.Fatalf("NewCache: %v", err)
	}
	defer instanceB.Close()

	ctx := context.Background()
	if err := instanceA.Scope("session").Set(ctx, "token", "alice", time.Minute); err != nil {
		t.Fatalf("Set on instance A: %v", err)
	}

	got, ok, err := instanceB.Scope("session").Get(ctx, "token")
	if err != nil {
		t.Fatalf("Get on instance B: %v", err)
	}
	if !ok {
		t.Fatal("token written on instance A is not visible on instance B")
	}
	if got != "alice" {
		t.Errorf("instance B read %q, want %q", got, "alice")
	}
}

func flushDB(t *testing.T, url string) {
	t.Helper()
	opts, err := redis.ParseURL(url)
	if err != nil {
		t.Fatalf("ParseURL(%q): %v", url, err)
	}
	client := redis.NewClient(opts)
	defer client.Close()
	if err := client.FlushDB(context.Background()).Err(); err != nil {
		t.Fatalf("FlushDB: %v", err)
	}
}
