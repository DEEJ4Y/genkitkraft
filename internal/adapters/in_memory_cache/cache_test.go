package inmemorycache_test

import (
	"bytes"
	"context"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"

	inmemorycache "github.com/DEEJ4Y/genkitkraft/internal/adapters/in_memory_cache"
	"github.com/DEEJ4Y/genkitkraft/internal/ports/cache"
	"github.com/DEEJ4Y/genkitkraft/resources/test/cachecontract"
)

func TestInMemoryCache(t *testing.T) {
	cachecontract.Run(t, func(t *testing.T) cache.Store {
		return inmemorycache.NewCache(time.Minute, zerolog.New(io.Discard))
	})
}

// A non-counter under a counter key means either a bug or another writer on the
// keyspace. Healing it is right, but healing it silently destroys the only
// evidence it ever happened — the whole point of the reset being observable.
func TestIncrementLogsWhenHealingANonCounter(t *testing.T) {
	var buf bytes.Buffer
	c := inmemorycache.NewCache(time.Minute, zerolog.New(&buf)).Scope("rate_limit")
	ctx := context.Background()

	if err := c.Set(ctx, "1.2.3.4", "not-a-number", time.Minute); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if _, err := c.Increment(ctx, "1.2.3.4", time.Minute); err != nil {
		t.Fatalf("Increment: %v", err)
	}

	logged := buf.String()
	if !strings.Contains(logged, "non-counter") {
		t.Errorf("heal was not logged; log = %q", logged)
	}
	// The fully-qualified key is what tells an operator which namespace was written
	// to, which is the first thing they need to find the offending writer.
	if !strings.Contains(logged, "rate_limit:1.2.3.4") {
		t.Errorf("heal log does not name the fully-qualified key; log = %q", logged)
	}
}

// The counterpart to the test above, and the one that keeps the heal log worth
// reading. Increment creates the key on the first attempt from every client, so
// logging that path would Warn on every first login from every IP and bury the
// real signal in noise.
func TestIncrementDoesNotLogWhenCreatingACounter(t *testing.T) {
	var buf bytes.Buffer
	c := inmemorycache.NewCache(time.Minute, zerolog.New(&buf)).Scope("rate_limit")

	if _, err := c.Increment(context.Background(), "1.2.3.4", time.Minute); err != nil {
		t.Fatalf("Increment: %v", err)
	}

	if buf.Len() != 0 {
		t.Errorf("creating a counter logged %q, want nothing", buf.String())
	}
}

// An expired counter reaches the same reset path as a corrupt one, and is just as
// ordinary as a first create: windows are meant to lapse.
func TestIncrementDoesNotLogWhenCounterExpired(t *testing.T) {
	var buf bytes.Buffer
	c := inmemorycache.NewCache(time.Millisecond, zerolog.New(&buf)).Scope("rate_limit")
	ctx := context.Background()

	if _, err := c.Increment(ctx, "1.2.3.4", 50*time.Millisecond); err != nil {
		t.Fatalf("Increment: %v", err)
	}
	// A fixed sleep is safe here because it is one-directional: oversleeping only
	// makes expiry more certain. Polling would not work — each poll is an Increment,
	// which would recreate the very counter being waited on.
	time.Sleep(150 * time.Millisecond)

	n, err := c.Increment(ctx, "1.2.3.4", 50*time.Millisecond)
	if err != nil {
		t.Fatalf("Increment after expiry: %v", err)
	}
	if n != 1 {
		t.Errorf("Increment after expiry = %d, want 1 (a fresh window)", n)
	}
	if buf.Len() != 0 {
		t.Errorf("an expired counter logged %q, want nothing", buf.String())
	}
}

// Increment must not lose counts under concurrent callers — the rate limit
// depends on every attempt being observed exactly once.
func TestIncrementIsAtomicUnderConcurrency(t *testing.T) {
	c := inmemorycache.NewCache(time.Minute, zerolog.New(io.Discard)).Scope("ns")
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
	c := inmemorycache.NewCache(time.Millisecond, zerolog.New(io.Discard)).Scope("ns")
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
