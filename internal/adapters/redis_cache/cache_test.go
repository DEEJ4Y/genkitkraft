//go:build integration

package rediscache_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

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

		store, err := rediscache.NewCache(url)
		if err != nil {
			t.Fatalf("NewCache: %v", err)
		}
		t.Cleanup(func() { store.Close() })
		return store
	})
}

// Valkey deployments are commonly addressed with a valkey:// URL, which go-redis
// does not recognise — the adapter must accept it.
func TestValkeySchemeIsAccepted(t *testing.T) {
	url := containers.StartValkeyURL(t)
	valkeyURL := "valkey://" + strings.TrimPrefix(url, "redis://")

	store, err := rediscache.NewCache(valkeyURL)
	if err != nil {
		t.Fatalf("NewCache(%q): %v", valkeyURL, err)
	}
	defer store.Close()

	c := store.Scope("ns")
	ctx := context.Background()
	if err := c.Set(ctx, "key", "value", time.Minute); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if got, ok := c.Get(ctx, "key"); !ok || got != "value" {
		t.Errorf("Get = (%q, %v), want (%q, true)", got, ok, "value")
	}
}

// A bad URL must fail at construction, so a misconfigured deployment dies at boot
// rather than serving 401s once traffic arrives.
func TestNewCacheFailsOnUnreachableServer(t *testing.T) {
	// Port 1 is reserved and never listening.
	if _, err := rediscache.NewCache("redis://127.0.0.1:1"); err == nil {
		t.Error("NewCache against an unreachable server: want error, got nil")
	}
}

func TestNewCacheFailsOnMalformedURL(t *testing.T) {
	if _, err := rediscache.NewCache("not-a-url"); err == nil {
		t.Error("NewCache with a malformed URL: want error, got nil")
	}
}

// Two Cache instances against one server model two application instances; the
// point of the Redis adapter is that they see each other's writes.
func TestSeparateInstancesShareState(t *testing.T) {
	url := containers.StartRedisURL(t)
	flushDB(t, url)

	instanceA, err := rediscache.NewCache(url)
	if err != nil {
		t.Fatalf("NewCache: %v", err)
	}
	defer instanceA.Close()

	instanceB, err := rediscache.NewCache(url)
	if err != nil {
		t.Fatalf("NewCache: %v", err)
	}
	defer instanceB.Close()

	ctx := context.Background()
	if err := instanceA.Scope("session").Set(ctx, "token", "alice", time.Minute); err != nil {
		t.Fatalf("Set on instance A: %v", err)
	}

	got, ok := instanceB.Scope("session").Get(ctx, "token")
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
