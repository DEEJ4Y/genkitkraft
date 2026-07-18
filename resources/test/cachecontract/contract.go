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
		got, ok, err := c.Get(ctx, "key")
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if !ok {
			t.Fatal("Get: want found, got miss")
		}
		if got != "value" {
			t.Errorf("Get = %q, want %q", got, "value")
		}
	})

	// A miss and an outage are different facts. An adapter reporting an absent key as
	// an error would make every caller read "not cached" as a dependency failure;
	// one reporting an outage as a miss is the bug this error return exists to
	// prevent, and only a networked adapter can exhibit it.
	t.Run("GetMissing", func(t *testing.T) {
		c := newStore(t).Scope("ns")
		_, ok, err := c.Get(context.Background(), "absent")
		if err != nil {
			t.Fatalf("Get on an absent key: want a nil error, got %v", err)
		}
		if ok {
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
		if _, ok, err := c.Get(ctx, "key"); err != nil {
			t.Fatalf("Get after Delete: %v", err)
		} else if ok {
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
		if _, ok, err := c.Get(ctx, "key"); err != nil {
			t.Fatalf("Get before TTL: %v", err)
		} else if !ok {
			t.Fatal("Get before TTL: want found, got miss")
		}
		if !eventually(t, 2*time.Second, func() bool {
			_, ok, err := c.Get(ctx, "key")
			return err == nil && !ok
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
		if _, ok, err := b.Get(ctx, "key"); err != nil {
			t.Fatalf("Get: %v", err)
		} else if ok {
			t.Error("key written in one namespace is visible in another")
		}

		if err := b.Set(ctx, "key", "from-beta", time.Minute); err != nil {
			t.Fatalf("Set: %v", err)
		}
		got, ok, err := a.Get(ctx, "key")
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
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

	// Asserting only a nil error is not enough: it passes on a backend whose
	// Decrement creates the key (Redis DECR creates it at -1), which both invents a
	// counter and leaves it with no expiry — a key that never expires, one per
	// affected IP. The absence has to be asserted directly.
	t.Run("DecrementAbsentIsNoOp", func(t *testing.T) {
		c := newStore(t).Scope("ns")
		ctx := context.Background()

		if err := c.Decrement(ctx, "absent"); err != nil {
			t.Errorf("Decrement on absent key: %v", err)
		}
		if _, ok, err := c.Get(ctx, "absent"); err != nil {
			t.Fatalf("Get: %v", err)
		} else if ok {
			t.Error("Decrement on an absent key created it")
		}
		n, err := c.Increment(ctx, "absent", time.Minute)
		if err != nil {
			t.Fatalf("Increment: %v", err)
		}
		if n != 1 {
			t.Errorf("Increment after Decrement on an absent key = %d, want 1", n)
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

	// Counters and strings share a keyspace. Redis cannot distinguish them — INCR
	// and SET both produce a string — so the decimal form is the only behaviour
	// both adapters can offer. Leaving it unspecified would also make "Get missed"
	// unreliable as proof that a key is absent, which DecrementAbsentIsNoOp needs.
	t.Run("GetOnCounterReturnsDecimalString", func(t *testing.T) {
		c := newStore(t).Scope("ns")
		ctx := context.Background()

		for i := 0; i < 3; i++ {
			if _, err := c.Increment(ctx, "counter", time.Minute); err != nil {
				t.Fatalf("Increment: %v", err)
			}
		}
		got, ok, err := c.Get(ctx, "counter")
		if err != nil {
			t.Fatalf("Get on a counter: %v", err)
		}
		if !ok {
			t.Fatal("Get on a counter: want found, got miss")
		}
		if got != "3" {
			t.Errorf("Get on a counter = %q, want %q", got, "3")
		}
	})

	// A zero TTL means "no expiration" (see cache.Cache). Redis deletes a key given
	// a non-positive expiry, so a PEXPIRE applied unconditionally would drop the
	// counter on every create — pinning it at 1 and stopping a rate limit built on
	// this from ever firing.
	t.Run("IncrementWithZeroTTLDoesNotExpire", func(t *testing.T) {
		c := newStore(t).Scope("ns")
		ctx := context.Background()

		for i := int64(1); i <= 3; i++ {
			n, err := c.Increment(ctx, "counter", 0)
			if err != nil {
				t.Fatalf("Increment %d: %v", i, err)
			}
			if n != i {
				t.Fatalf("Increment #%d with a zero TTL = %d, want %d", i, n, i)
			}
		}
		// A fixed sleep is safe here: the assertion is that the counter SURVIVES, so
		// a slow machine only makes survival more certain. Do not turn this into an
		// eventually() poll — the failure mode is one-directional.
		time.Sleep(200 * time.Millisecond)
		n, err := c.Increment(ctx, "counter", 0)
		if err != nil {
			t.Fatalf("Increment: %v", err)
		}
		if n != 4 {
			t.Errorf("Increment after a pause with a zero TTL = %d, want 4 (the counter expired)", n)
		}
	})

	// A non-counter under a counter key can only come from a bug or an outside
	// writer, but it has to be recoverable: on Redis a leftover string may carry no
	// TTL, so nothing would ever clear it and the caller would be stuck for good.
	t.Run("IncrementRecoversFromNonCounterValue", func(t *testing.T) {
		c := newStore(t).Scope("ns")
		ctx := context.Background()

		if err := c.Set(ctx, "counter", "not-a-number", time.Minute); err != nil {
			t.Fatalf("Set: %v", err)
		}
		n, err := c.Increment(ctx, "counter", time.Minute)
		if err != nil {
			t.Fatalf("Increment over a non-counter value: %v", err)
		}
		if n != 1 {
			t.Errorf("Increment over a non-counter value = %d, want 1 (a fresh counter)", n)
		}
		n, err = c.Increment(ctx, "counter", time.Minute)
		if err != nil {
			t.Fatalf("Increment: %v", err)
		}
		if n != 2 {
			t.Errorf("second Increment after recovery = %d, want 2", n)
		}
	})

	// The replacement counter must carry the TTL, or healing a corrupt key would
	// just trade a stuck value for one that never expires.
	t.Run("IncrementRecoveryAppliesTTL", func(t *testing.T) {
		c := newStore(t).Scope("ns")
		ctx := context.Background()

		if err := c.Set(ctx, "counter", "not-a-number", time.Minute); err != nil {
			t.Fatalf("Set: %v", err)
		}
		if _, err := c.Increment(ctx, "counter", 100*time.Millisecond); err != nil {
			t.Fatalf("Increment: %v", err)
		}
		if !eventually(t, 2*time.Second, func() bool {
			n, err := c.Increment(ctx, "counter", 100*time.Millisecond)
			return err == nil && n == 1
		}) {
			t.Error("counter healed from a non-counter value never expired; the TTL was not applied on reset")
		}
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
