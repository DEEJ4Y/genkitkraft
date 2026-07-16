// Package rediscache implements the cache port against Redis, and against any
// wire-compatible server such as Valkey. Both speak the same protocol and
// command set, so a single adapter serves both — only the URL scheme differs,
// which normalizeURL handles.
package rediscache

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/DEEJ4Y/genkitkraft/internal/ports/cache"
)

var _ cache.Store = (*Cache)(nil)

// pingTimeout bounds the startup connectivity check so a wrong CACHE_URL fails
// fast at boot rather than on the first authenticated request.
const pingTimeout = 5 * time.Second

// incrementScript increments a counter and applies the TTL only when the key is
// created. INCR and PEXPIRE must run as one unit: as two round trips, a crash or
// a lost connection between them leaves a counter with no expiry, which would
// lock an IP out permanently.
var incrementScript = redis.NewScript(`
local n = redis.call('INCR', KEYS[1])
if n == 1 then
  redis.call('PEXPIRE', KEYS[1], ARGV[1])
end
return n
`)

// Cache is the Redis-backed backing store. It does not implement cache.Cache
// directly; use Scope to obtain a namespace-isolated cache.Cache.
type Cache struct {
	client *redis.Client
}

// NewCache connects to the Redis/Valkey server at rawURL and verifies the
// connection before returning.
func NewCache(rawURL string) (*Cache, error) {
	opts, err := redis.ParseURL(normalizeURL(rawURL))
	if err != nil {
		return nil, fmt.Errorf("parsing cache URL: %w", err)
	}

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("pinging cache: %w", err)
	}

	return &Cache{client: client}, nil
}

// normalizeURL maps the Valkey URL schemes onto their Redis equivalents.
// go-redis only recognises redis:// and rediss://, but operators configuring a
// Valkey server reasonably reach for valkey://.
func normalizeURL(rawURL string) string {
	switch {
	case strings.HasPrefix(rawURL, "valkeys://"):
		return "rediss://" + strings.TrimPrefix(rawURL, "valkeys://")
	case strings.HasPrefix(rawURL, "valkey://"):
		return "redis://" + strings.TrimPrefix(rawURL, "valkey://")
	default:
		return rawURL
	}
}

// Close releases the underlying connection pool.
func (c *Cache) Close() error {
	return c.client.Close()
}

// Scope returns a Cache where every key is prefixed with "<namespace>:".
func (c *Cache) Scope(namespace string) cache.Cache {
	return cache.Scope(&rawCache{client: c.client}, namespace)
}

// rawCache is unexported — it implements cache.Cache and is only used via Scope.
type rawCache struct {
	client *redis.Client
}

// Get reports a miss for any error, including a connection failure. The port
// signature has no error return, so an outage is indistinguishable from an
// expired key here: sessions fail closed and the user is asked to log in again.
func (r *rawCache) Get(ctx context.Context, key string) (string, bool) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return "", false
	}
	return val, true
}

func (r *rawCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	// go-redis treats a zero expiration as "no expiry", matching the port contract.
	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *rawCache) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *rawCache) Increment(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	n, err := incrementScript.Run(ctx, r.client, []string{key}, ttl.Milliseconds()).Int64()
	if err != nil {
		return 0, fmt.Errorf("cache: incrementing %q: %w", key, err)
	}
	return n, nil
}

func (r *rawCache) Decrement(ctx context.Context, key string) error {
	return r.client.Decr(ctx, key).Err()
}
