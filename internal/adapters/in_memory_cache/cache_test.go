package inmemorycache_test

import (
	"context"
	"sync"
	"testing"
	"time"

	inmemorycache "github.com/DEEJ4Y/genkitkraft/internal/adapters/in_memory_cache"
	"github.com/DEEJ4Y/genkitkraft/internal/ports/cache"
	"github.com/DEEJ4Y/genkitkraft/resources/test/cachecontract"
)

func TestInMemoryCache(t *testing.T) {
	cachecontract.Run(t, func(t *testing.T) cache.Store {
		return inmemorycache.NewCache(time.Minute)
	})
}

// Increment must not lose counts under concurrent callers — the rate limit
// depends on every attempt being observed exactly once.
func TestIncrementIsAtomicUnderConcurrency(t *testing.T) {
	c := inmemorycache.NewCache(time.Minute).Scope("ns")
	ctx := context.Background()

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			if _, err := c.Increment(ctx, "counter", time.Minute); err != nil {
				t.Errorf("Increment: %v", err)
			}
		}()
	}
	wg.Wait()

	n, err := c.Increment(ctx, "counter", time.Minute)
	if err != nil {
		t.Fatalf("Increment: %v", err)
	}
	if want := int64(goroutines + 1); n != want {
		t.Errorf("counter after %d concurrent increments = %d, want %d", goroutines, n, want)
	}
}

// A counter and a string can collide on one key only through caller error, but
// Increment must report that rather than silently resetting the count.
func TestIncrementOnStringKeyErrors(t *testing.T) {
	c := inmemorycache.NewCache(time.Minute).Scope("ns")
	ctx := context.Background()

	if err := c.Set(ctx, "key", "not-a-number", time.Minute); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if _, err := c.Increment(ctx, "key", time.Minute); err == nil {
		t.Error("Increment on a string key: want error, got nil")
	}
}
