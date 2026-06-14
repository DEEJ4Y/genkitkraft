package cache

import (
	"context"
	"time"
)

// Cache defines the contract for key-value caching with TTL support.
type Cache interface {
	// Get retrieves a value by key. Returns the value and whether it was found and unexpired.
	Get(ctx context.Context, key string) (string, bool)
	// Set stores a value with the given TTL. A zero TTL means no expiration.
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	// Delete removes the entry for key. No-op if the key does not exist.
	Delete(ctx context.Context, key string) error
}

// Store is the factory for obtaining namespace-isolated Cache instances.
// Adapters implement Store rather than Cache directly — callers must always call Scope.
type Store interface {
	Scope(namespace string) Cache
}
