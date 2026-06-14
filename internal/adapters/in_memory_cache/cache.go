package inmemorychache

import (
	"context"
	"time"

	gocache "github.com/patrickmn/go-cache"

	"github.com/DEEJ4Y/genkitkraft/internal/ports/cache"
)

var _ cache.Store = (*Cache)(nil)

// Cache is the in-memory backing store. It does not implement cache.Cache directly;
// use Scope to obtain a namespace-isolated cache.Cache.
type Cache struct {
	store *gocache.Cache
}

// NewCache creates a new in-memory backing store.
// defaultExpiration is used when Set is called with a zero TTL.
// cleanupInterval controls how often expired entries are purged from memory.
func NewCache(defaultExpiration, cleanupInterval time.Duration) *Cache {
	return &Cache{store: gocache.New(defaultExpiration, cleanupInterval)}
}

// Scope returns a Cache where every key is prefixed with "<namespace>:".
func (c *Cache) Scope(namespace string) cache.Cache {
	return cache.Scope(&rawCache{store: c.store}, namespace)
}

// rawCache is unexported — it implements cache.Cache and is only used via Scope.
type rawCache struct {
	store *gocache.Cache
}

func (r *rawCache) Get(_ context.Context, key string) (string, bool) {
	val, found := r.store.Get(key)
	if !found {
		return "", false
	}
	str, ok := val.(string)
	return str, ok
}

func (r *rawCache) Set(_ context.Context, key string, value string, ttl time.Duration) error {
	r.store.Set(key, value, ttl)
	return nil
}
