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

// Increment must not error at a window boundary. The old check-then-increment
// left a gap: counterMu serializes callers against each other but not against
// go-cache's time-based expiry, so an entry could lapse between the found-check
// and the increment, which then reported "not found" — and the rate limiter turns
// any error into a denied login for a legitimate user.
//
// The design here is load-bearing, so do not "simplify" it into a fixed number of
// iterations. What exposes the race is the number of EXPIRY events, not the number
// of increments: the vulnerable gap is the ~100ns between the two calls, so a hit
// needs an expiry to land inside it. A fixed 10,000 increments finish in ~10ms and
// so produce only ~10 expiries — far too few, and such a test passes against the
// racy implementation. Running for a fixed duration against a very short TTL
// produces tens of thousands of expiries instead.
//
// Verified by mutation: against the old check-then-increment this observes ~270
// errors per second; against the current implementation it observes zero.
func TestIncrementNeverErrorsAtTheWindowBoundary(t *testing.T) {
	c := inmemorycache.NewCache(time.Millisecond).Scope("ns")
	ctx := context.Background()

	const (
		ttl      = 100 * time.Microsecond
		duration = 500 * time.Millisecond
	)

	deadline := time.Now().Add(duration)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for time.Now().Before(deadline) {
				if _, err := c.Increment(ctx, "counter", ttl); err != nil {
					t.Errorf("Increment at the window boundary: %v", err)
					return
				}
			}
		}()
	}
	wg.Wait()
}
