package memorysession

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	"github.com/DEEJ4Y/genkitkraft/internal/ports/cache"
	"github.com/DEEJ4Y/genkitkraft/internal/ports/session"
)

const sessionTTL = 24 * time.Hour

// Compile-time check that MemoryStore implements session.Store.
var _ session.Store = (*MemoryStore)(nil)

// MemoryStore manages sessions via the cache port. TTL-based expiry replaces the
// manual cleanup loop that the previous map-based implementation required.
type MemoryStore struct {
	cache cache.Cache
}

func NewMemoryStore(c cache.Cache) *MemoryStore {
	return &MemoryStore{cache: c}
}

func (s *MemoryStore) Create(ctx context.Context, username string) (string, error) {
	token := uuid.New().String()
	if err := s.cache.Set(ctx, token, username, sessionTTL); err != nil {
		return "", err
	}
	return token, nil
}

func (s *MemoryStore) Validate(ctx context.Context, token string) (string, error) {
	username, ok := s.cache.Get(ctx, token)
	if !ok {
		return "", errors.NewAppError(errors.Unauthorized, "invalid or expired session")
	}
	return username, nil
}

func (s *MemoryStore) Delete(ctx context.Context, token string) error {
	return s.cache.Delete(ctx, token)
}
