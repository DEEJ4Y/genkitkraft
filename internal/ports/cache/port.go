package cache

import (
	"context"
	"time"
)

// Cache defines the contract for key-value caching with TTL support.
type Cache interface {
	// Get retrieves a value by key. It reports whether the key was found and
	// unexpired, and returns a non-nil error only when the backing store could not
	// answer at all — a connection failure, say.
	//
	// A miss and an outage are different facts and must not be conflated: found is
	// false for both, but only an error means "we do not know". A caller that fails
	// closed on a miss — session validation above all — would otherwise report a
	// dependency outage as an expired session. Check the error before the flag.
	Get(ctx context.Context, key string) (string, bool, error)
	// Set stores a value with the given TTL. A zero TTL means no expiration.
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	// Delete removes the entry for key. No-op if the key does not exist.
	Delete(ctx context.Context, key string) error
	// Increment atomically increments the integer counter at key and returns the
	// new value. If key does not exist it is created with value 1 and the given
	// TTL. The TTL is applied only on creation — later increments leave it
	// untouched, so the counter expires a fixed window after the first increment
	// rather than sliding forward on every hit.
	//
	// Implementations must be atomic across all callers sharing the backing
	// store, including separate processes, so this can back a distributed
	// rate limit.
	Increment(ctx context.Context, key string, ttl time.Duration) (int64, error)
	// Decrement atomically decrements the integer counter at key, leaving its TTL
	// untouched. No-op if the key does not exist.
	Decrement(ctx context.Context, key string) error
}

// Store is the factory for obtaining namespace-isolated Cache instances.
// Adapters implement Store rather than Cache directly — callers must always call Scope.
type Store interface {
	Scope(namespace string) Cache
}
