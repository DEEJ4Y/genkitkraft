// Package cachecontract holds the conformance suite every cache.Store adapter
// must pass. Sessions and login rate limiting are swapped between the in-memory
// and Redis adapters purely by configuration, so the two must be observably
// interchangeable — running one suite against both is what enforces that.
package cachecontract

import (
	"context"
	"testing"
	"time"

	"github.com/DEEJ4Y/genkitkraft/internal/ports/cache"
)

// Run exercises store against the full cache.Store contract.
// newStore must return a distinct, empty store on each call so subtests do not
// share keys.
func Run(t *testing.T, newStore func(t *testing.T) cache.Store) {
	t.Helper()

	t.Run("SetThenGet", func(t *testing.T) {
		c := newStore(t).Scope("ns")
		ctx := context.Background()

		if err := c.Set(ctx, "key", "value", time.Minute); err != nil {
			t.Fatalf("Set: %v", err)
		}
		got, ok := c.Get(ctx, "key")
		if !ok {
			t.Fatal("Get: want found, got miss")
		}
		if got != "value" {
			t.Errorf("Get = %q, want %q", got, "value")
		}
	})

	t.Run("GetMissing", func(t *testing.T) {
		c := newStore(t).Scope("ns")
		if _, ok := c.Get(context.Background(), "absent"); ok {
			t.Error("Get on absent key: want miss, got found")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		c := newStore(t).Scope("ns")
		ctx := context.Background()

		if err := c.Set(ctx, "key", "value", time.Minute); err != nil {
			t.Fatalf("Set: %v", err)
		}
		if err := c.Delete(ctx, "key"); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		if _, ok := c.Get(ctx, "key"); ok {
			t.Error("Get after Delete: want miss, got found")
		}
	})

	t.Run("DeleteAbsentIsNoOp", func(t *testing.T) {
		c := newStore(t).Scope("ns")
		if err := c.Delete(context.Background(), "absent"); err != nil {
			t.Errorf("Delete on absent key: %v", err)
		}
	})

	t.Run("TTLExpires", func(t *testing.T) {
		c := newStore(t).Scope("ns")
		ctx := context.Background()

		if err := c.Set(ctx, "key", "value", 100*time.Millisecond); err != nil {
			t.Fatalf("Set: %v", err)
		}
		if _, ok := c.Get(ctx, "key"); !ok {
			t.Fatal("Get before TTL: want found, got miss")
		}
		if !eventually(t, 2*time.Second, func() bool {
			_, ok := c.Get(ctx, "key")
			return !ok
		}) {
			t.Error("key still present well after its TTL elapsed")
		}
	})

	t.Run("NamespacesAreIsolated", func(t *testing.T) {
		store := newStore(t)
		a, b := store.Scope("alpha"), store.Scope("beta")
		ctx := context.Background()

		if err := a.Set(ctx, "key", "from-alpha", time.Minute); err != nil {
			t.Fatalf("Set: %v", err)
		}
		if _, ok := b.Get(ctx, "key"); ok {
			t.Error("key written in one namespace is visible in another")
		}

		if err := b.Set(ctx, "key", "from-beta", time.Minute); err != nil {
			t.Fatalf("Set: %v", err)
		}
		got, ok := a.Get(ctx, "key")
		if !ok || got != "from-alpha" {
			t.Errorf("alpha key = (%q, %v) after beta wrote the same key, want (%q, true)", got, ok, "from-alpha")
		}
	})

	t.Run("IncrementCreatesAtOne", func(t *testing.T) {
		c := newStore(t).Scope("ns")
		n, err := c.Increment(context.Background(), "counter", time.Minute)
		if err != nil {
			t.Fatalf("Increment: %v", err)
		}
		if n != 1 {
			t.Errorf("first Increment = %d, want 1", n)
		}
	})

	t.Run("IncrementAccumulates", func(t *testing.T) {
		c := newStore(t).Scope("ns")
		ctx := context.Background()

		for i := int64(1); i <= 5; i++ {
			n, err := c.Increment(ctx, "counter", time.Minute)
			if err != nil {
				t.Fatalf("Increment %d: %v", i, err)
			}
			if n != i {
				t.Fatalf("Increment #%d = %d, want %d", i, n, i)
			}
		}
	})

	t.Run("IncrementAppliesTTLOnCreate", func(t *testing.T) {
		c := newStore(t).Scope("ns")
		ctx := context.Background()

		if _, err := c.Increment(ctx, "counter", 100*time.Millisecond); err != nil {
			t.Fatalf("Increment: %v", err)
		}
		if !eventually(t, 2*time.Second, func() bool {
			n, err := c.Increment(ctx, "counter", 100*time.Millisecond)
			return err == nil && n == 1
		}) {
			t.Error("counter did not expire after its window; it never reset to 1")
		}
	})

	// The window must run from the first increment. If a later increment reset the
	// TTL, a steady stream of attempts would hold the counter alive indefinitely
	// and an IP could stay locked out long past the intended window.
	t.Run("IncrementDoesNotExtendTTL", func(t *testing.T) {
		c := newStore(t).Scope("ns")
		ctx := context.Background()

		if _, err := c.Increment(ctx, "counter", 300*time.Millisecond); err != nil {
			t.Fatalf("Increment: %v", err)
		}
		// Keep hitting the counter across most of the window.
		for i := 0; i < 4; i++ {
			time.Sleep(50 * time.Millisecond)
			if _, err := c.Increment(ctx, "counter", 300*time.Millisecond); err != nil {
				t.Fatalf("Increment: %v", err)
			}
		}
		if !eventually(t, 2*time.Second, func() bool {
			n, err := c.Increment(ctx, "counter", 300*time.Millisecond)
			return err == nil && n == 1
		}) {
			t.Error("counter outlived its original window, so increments extended the TTL")
		}
	})

	t.Run("Decrement", func(t *testing.T) {
		c := newStore(t).Scope("ns")
		ctx := context.Background()

		for i := 0; i < 3; i++ {
			if _, err := c.Increment(ctx, "counter", time.Minute); err != nil {
				t.Fatalf("Increment: %v", err)
			}
		}
		if err := c.Decrement(ctx, "counter"); err != nil {
			t.Fatalf("Decrement: %v", err)
		}
		n, err := c.Increment(ctx, "counter", time.Minute)
		if err != nil {
			t.Fatalf("Increment: %v", err)
		}
		if n != 3 {
			t.Errorf("Increment after 3x Increment + 1x Decrement = %d, want 3", n)
		}
	})

	t.Run("DecrementAbsentIsNoOp", func(t *testing.T) {
		c := newStore(t).Scope("ns")
		if err := c.Decrement(context.Background(), "absent"); err != nil {
			t.Errorf("Decrement on absent key: %v", err)
		}
	})

	t.Run("DeleteResetsCounter", func(t *testing.T) {
		c := newStore(t).Scope("ns")
		ctx := context.Background()

		for i := 0; i < 3; i++ {
			if _, err := c.Increment(ctx, "counter", time.Minute); err != nil {
				t.Fatalf("Increment: %v", err)
			}
		}
		if err := c.Delete(ctx, "counter"); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		n, err := c.Increment(ctx, "counter", time.Minute)
		if err != nil {
			t.Fatalf("Increment: %v", err)
		}
		if n != 1 {
			t.Errorf("Increment after Delete = %d, want 1", n)
		}
	})

	// Counters and strings share a keyspace; each adapter must keep Get's contract
	// (a string or a miss) rather than panicking on a counter value.
	t.Run("GetOnCounterDoesNotPanic", func(t *testing.T) {
		c := newStore(t).Scope("ns")
		ctx := context.Background()

		if _, err := c.Increment(ctx, "counter", time.Minute); err != nil {
			t.Fatalf("Increment: %v", err)
		}
		c.Get(ctx, "counter") // value is unspecified; must not panic
	})
}

// eventually polls cond until it holds or timeout elapses. TTL expiry is not
// instantaneous on either backend, so polling avoids a fixed sleep that is both
// slow and flaky.
func eventually(t *testing.T, timeout time.Duration, cond func() bool) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return true
		}
		time.Sleep(20 * time.Millisecond)
	}
	return false
}
