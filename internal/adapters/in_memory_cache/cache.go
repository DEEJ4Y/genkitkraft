package inmemorycache

import (
	"context"
	"fmt"
	"sync"
	"time"

	gocache "github.com/patrickmn/go-cache"

	"github.com/DEEJ4Y/genkitkraft/internal/ports/cache"
)

var _ cache.Store = (*Cache)(nil)

// Cache is the in-memory backing store. It does not implement cache.Cache directly;
// use Scope to obtain a namespace-isolated cache.Cache.
type Cache struct {
	store *gocache.Cache
	// counterMu serializes the read-modify-write in Increment/Decrement. go-cache
	// locks each individual operation but exposes no compare-and-set, so the
	// not-found-then-create path in Increment needs its own lock to stay atomic.
	counterMu sync.Mutex
}

// NewCache creates a new in-memory backing store.
// cleanupInterval controls how often expired entries are purged from memory.
// TTLs are always specified per-Set call; there is no global default expiration.
//
// NOTE: This store is process-local and has no maximum-entry cap. It is the right
// choice for a single-instance deployment only. Running multiple instances against
// it gives each its own sessions and its own rate-limit counters — set
// CACHE_PROVIDER=redis (or valkey) to share state across instances.
func NewCache(cleanupInterval time.Duration) *Cache {
	return &Cache{store: gocache.New(gocache.NoExpiration, cleanupInterval)}
}

// Scope returns a Cache where every key is prefixed with "<namespace>:".
func (c *Cache) Scope(namespace string) cache.Cache {
	return cache.Scope(&rawCache{owner: c}, namespace)
}

// rawCache is unexported — it implements cache.Cache and is only used via Scope.
type rawCache struct {
	owner *Cache
}

func (r *rawCache) Get(_ context.Context, key string) (string, bool) {
	val, found := r.owner.store.Get(key)
	if !found {
		return "", false
	}
	str, ok := val.(string)
	return str, ok
}

func (r *rawCache) Set(_ context.Context, key string, value string, ttl time.Duration) error {
	r.owner.store.Set(key, value, ttl)
	return nil
}

func (r *rawCache) Delete(_ context.Context, key string) error {
	r.owner.store.Delete(key)
	return nil
}

func (r *rawCache) Increment(_ context.Context, key string, ttl time.Duration) (int64, error) {
	r.owner.counterMu.Lock()
	defer r.owner.counterMu.Unlock()

	// IncrementInt64 mutates the value in place, so the entry keeps the expiry it
	// was created with. It errors when the key is absent, expired, or holds a
	// non-int64 — distinguish the last case, since that is a real bug rather than
	// a first increment.
	if existing, found := r.owner.store.Get(key); found {
		if _, ok := existing.(int64); !ok {
			return 0, fmt.Errorf("cache: key %q holds %T, not a counter", key, existing)
		}
		n, err := r.owner.store.IncrementInt64(key, 1)
		if err != nil {
			return 0, fmt.Errorf("cache: incrementing %q: %w", key, err)
		}
		return n, nil
	}

	r.owner.store.Set(key, int64(1), ttl)
	return 1, nil
}

func (r *rawCache) Decrement(_ context.Context, key string) error {
	r.owner.counterMu.Lock()
	defer r.owner.counterMu.Unlock()

	// A missing key means the window already expired; there is nothing to undo.
	if _, found := r.owner.store.Get(key); !found {
		return nil
	}
	if _, err := r.owner.store.DecrementInt64(key, 1); err != nil {
		return fmt.Errorf("cache: decrementing %q: %w", key, err)
	}
	return nil
}
