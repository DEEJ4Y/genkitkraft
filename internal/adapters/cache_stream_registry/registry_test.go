package cachestreamregistry_test

import (
	"context"
	"io"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rs/zerolog"

	cachestreamregistry "github.com/DEEJ4Y/genkitkraft/internal/adapters/cache_stream_registry"
	inmemorycache "github.com/DEEJ4Y/genkitkraft/internal/adapters/in_memory_cache"
	"github.com/DEEJ4Y/genkitkraft/internal/ports/cache"
)

const testPollInterval = 20 * time.Millisecond

func newStore(t *testing.T) cache.Store {
	t.Helper()
	return inmemorycache.NewCache(time.Minute, zerolog.New(io.Discard))
}

func newRegistry(t *testing.T, c cache.Cache) *cachestreamregistry.Registry {
	t.Helper()
	return cachestreamregistry.NewRegistry(c, testPollInterval, zerolog.New(io.Discard))
}

func TestCancelInvokesRegisteredFunc(t *testing.T) {
	r := newRegistry(t, newStore(t).Scope("ns"))

	var called atomic.Bool
	_, cancel := context.WithCancel(context.Background())
	r.Register("msg-1", func() { called.Store(true); cancel() })

	if !r.Cancel("msg-1") {
		t.Fatal("Cancel on a registered ID: want true, got false")
	}
	if !called.Load() {
		t.Error("Cancel did not invoke the registered cancel function")
	}
}

func TestCancelOnUnknownIDReturnsFalseButStillSignals(t *testing.T) {
	store := newStore(t)
	c := store.Scope("ns")
	r := newRegistry(t, c)

	if r.Cancel("never-registered") {
		t.Error("Cancel on an unknown ID: want false, got true")
	}

	_, ok, err := c.Get(context.Background(), "never-registered")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !ok {
		t.Error("Cancel on an unknown ID did not write a cache signal")
	}
}

func TestUnregisterStopsPollingAndClearsSignal(t *testing.T) {
	store := newStore(t)
	c := store.Scope("ns")
	r := newRegistry(t, c)

	var called atomic.Bool
	r.Register("msg-1", func() { called.Store(true) })
	r.Unregister("msg-1")

	if _, ok, err := c.Get(context.Background(), "msg-1"); err != nil {
		t.Fatalf("Get: %v", err)
	} else if ok {
		t.Error("Unregister did not clear the cache signal")
	}

	// Plant a signal after Unregister and confirm the (stopped) poller never
	// reacts to it — proof the poller goroutine actually exited rather than
	// just happening not to have fired yet.
	if err := c.Set(context.Background(), "msg-1", "1", time.Minute); err != nil {
		t.Fatalf("Set: %v", err)
	}
	time.Sleep(5 * testPollInterval)
	if called.Load() {
		t.Error("a signal planted after Unregister triggered the cancel func; poller did not stop")
	}
}

func TestCrossInstanceCancelReachesTheOwningRegistry(t *testing.T) {
	store := newStore(t)

	registryA := newRegistry(t, store.Scope("stream_cancel"))
	registryB := newRegistry(t, store.Scope("stream_cancel"))

	done := make(chan struct{})
	registryA.Register("msg-1", func() { close(done) })

	if registryB.Cancel("msg-1") {
		t.Error("Cancel on the non-owning registry: want false (it has no local registration), got true")
	}

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("registry A's cancel func was never invoked after registry B's Cancel")
	}
}

func TestDoubleCancelInvokesFuncExactlyOnce(t *testing.T) {
	r := newRegistry(t, newStore(t).Scope("ns"))

	var calls atomic.Int32
	r.Register("msg-1", func() { calls.Add(1) })

	r.Cancel("msg-1")
	r.Cancel("msg-1")

	if n := calls.Load(); n != 1 {
		t.Errorf("cancel func invoked %d times, want 1", n)
	}
}

func TestRegisterOverwritesPriorRegistration(t *testing.T) {
	store := newStore(t)
	c := store.Scope("ns")
	r := newRegistry(t, c)

	var firstCalled, secondCalled atomic.Bool
	r.Register("msg-1", func() { firstCalled.Store(true) })
	r.Register("msg-1", func() { secondCalled.Store(true) })

	r.Cancel("msg-1")

	if firstCalled.Load() {
		t.Error("the overwritten registration's cancel func was invoked")
	}
	if !secondCalled.Load() {
		t.Error("the current registration's cancel func was not invoked")
	}

	// The first registration's poller must also have stopped, not leaked.
	if err := c.Set(context.Background(), "msg-1", "1", time.Minute); err != nil {
		t.Fatalf("Set: %v", err)
	}
	time.Sleep(5 * testPollInterval)
	if firstCalled.Load() {
		t.Error("the overwritten registration's poller reacted to a later signal; it should have exited")
	}
}

// failingCache stands in for a cache backend that is down.
type failingCache struct{ err error }

func (f *failingCache) Get(context.Context, string) (string, bool, error) { return "", false, f.err }
func (f *failingCache) Set(context.Context, string, string, time.Duration) error {
	return f.err
}
func (f *failingCache) Delete(context.Context, string) error { return f.err }
func (f *failingCache) Increment(context.Context, string, time.Duration) (int64, error) {
	return 0, f.err
}
func (f *failingCache) Decrement(context.Context, string) error { return f.err }

func TestFailingCacheDoesNotBreakLocalCancellation(t *testing.T) {
	r := cachestreamregistry.NewRegistry(&failingCache{err: io.ErrUnexpectedEOF}, testPollInterval, zerolog.New(io.Discard))

	var called atomic.Bool
	r.Register("msg-1", func() { called.Store(true) })

	if !r.Cancel("msg-1") {
		t.Fatal("Cancel with a failing cache: want true (local cancellation still works), got false")
	}
	if !called.Load() {
		t.Error("local cancel func was not invoked despite the cache failing")
	}

	// Register/Unregister with a failing cache must not panic.
	r.Register("msg-2", func() {})
	r.Unregister("msg-2")
}
