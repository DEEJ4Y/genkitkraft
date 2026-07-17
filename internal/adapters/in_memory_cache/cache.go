package inmemorycache

import (
	"context"
	"fmt"
	"strconv"
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
	switch v := val.(type) {
	case string:
		return v, true
	case int64:
		// Counters read back as their decimal form. Redis keeps INCR counters and
		// SET strings in one string keyspace and cannot tell them apart, so matching
		// that here is what keeps the two adapters interchangeable — and it makes a
		// Get miss reliable proof that a key is absent.
		return strconv.FormatInt(v, 10), true
	default:
		return "", false
	}
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

	// Increment in place first. A found-check beforehand would be a TOCTOU:
	// counterMu serializes callers but not go-cache's time-based expiry, so an
	// entry can lapse between the check and the increment — and IncrementInt64 runs
	// its own expiry check, so it would then report "not found" and the limiter
	// would reject a legitimate attempt right at the window boundary.
	//
	// In-place mutation also means the entry keeps the expiry it was created with,
	// so the window runs from the first increment rather than sliding forward.
	if n, err := r.owner.store.IncrementInt64(key, 1); err == nil {
		return n, nil
	}
	// IncrementInt64 fails when the key is absent, expired, or holds a non-int64.
	// All three start a fresh window: for the first two that is the contract, and
	// for the last it is self-healing — a stuck non-counter has no way to clear
	// itself, so erroring would lock the IP out for good.
	r.owner.store.Set(key, int64(1), ttl)
	return 1, nil
}

func (r *rawCache) Decrement(_ context.Context, key string) error {
	r.owner.counterMu.Lock()
	defer r.owner.counterMu.Unlock()

	if _, err := r.owner.store.DecrementInt64(key, 1); err != nil {
		// DecrementInt64 fails when the key is absent, expired, or holds a non-int64.
		// A missing key means the window already expired and there is nothing to
		// undo. Check after the fact rather than before, so expiry cannot race the
		// check; if the key lapsed in between, the no-op is correct anyway.
		if _, found := r.owner.store.Get(key); !found {
			return nil
		}
		return fmt.Errorf("cache: decrementing %q: %w", key, err)
	}
	return nil
}
