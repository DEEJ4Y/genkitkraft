package cache

import (
	"context"
	"time"
)

type scopedCache struct {
	inner  Cache
	prefix string
}

// Scope wraps c so all keys are stored as "<namespace>:<key>".
// Adapters call this from their own Scope method to reuse this logic.
func Scope(c Cache, namespace string) Cache {
	return &scopedCache{inner: c, prefix: namespace + ":"}
}

func (s *scopedCache) Get(ctx context.Context, key string) (string, bool, error) {
	return s.inner.Get(ctx, s.prefix+key)
}

func (s *scopedCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return s.inner.Set(ctx, s.prefix+key, value, ttl)
}

func (s *scopedCache) Delete(ctx context.Context, key string) error {
	return s.inner.Delete(ctx, s.prefix+key)
}

func (s *scopedCache) Increment(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	return s.inner.Increment(ctx, s.prefix+key, ttl)
}

func (s *scopedCache) Decrement(ctx context.Context, key string) error {
	return s.inner.Decrement(ctx, s.prefix+key)
}
