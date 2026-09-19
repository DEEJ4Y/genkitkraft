//go:build integration

package cachestreamregistry_test

import (
	"io"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rs/zerolog"

	cachestreamregistry "github.com/DEEJ4Y/genkitkraft/internal/adapters/cache_stream_registry"
	rediscache "github.com/DEEJ4Y/genkitkraft/internal/adapters/redis_cache"
	"github.com/DEEJ4Y/genkitkraft/resources/test/containers"
)

// Two Registry instances backed by two independent Cache connections to one
// shared Redis/Valkey server model two application instances. This is the
// literal scenario the cache-backed registry exists for: a "stop generation"
// request landing on the instance that is not running the generation.
const integrationPollInterval = 50 * time.Millisecond

func TestCrossInstanceCancelOverRedis(t *testing.T) {
	testCrossInstanceCancel(t, containers.StartRedisURL(t))
}

func TestCrossInstanceCancelOverValkey(t *testing.T) {
	testCrossInstanceCancel(t, containers.StartValkeyURL(t))
}

func testCrossInstanceCancel(t *testing.T, url string) {
	t.Helper()

	storeA, err := rediscache.NewCache(url, zerolog.New(io.Discard))
	if err != nil {
		t.Fatalf("NewCache (instance A): %v", err)
	}
	defer storeA.Close()

	storeB, err := rediscache.NewCache(url, zerolog.New(io.Discard))
	if err != nil {
		t.Fatalf("NewCache (instance B): %v", err)
	}
	defer storeB.Close()

	registryA := cachestreamregistry.NewRegistry(storeA.Scope("stream_cancel"), integrationPollInterval, zerolog.New(io.Discard))
	registryB := cachestreamregistry.NewRegistry(storeB.Scope("stream_cancel"), integrationPollInterval, zerolog.New(io.Discard))

	var cancelled atomic.Bool
	registryA.Register("msg-1", func() { cancelled.Store(true) })

	if registryB.Cancel("msg-1") {
		t.Error("Cancel on the non-owning instance: want false, got true")
	}

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if cancelled.Load() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("instance A never noticed instance B's cancel signal")
}
