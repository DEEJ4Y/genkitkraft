package memorysession

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	"github.com/DEEJ4Y/genkitkraft/internal/ports/session"
)

const sessionTTL = 24 * time.Hour

// Compile-time check that MemoryStore implements session.Store.
var _ session.Store = (*MemoryStore)(nil)

type storedSession struct {
	Username  string
	ExpiresAt time.Time
}

// MemoryStore manages sessions in memory.
type MemoryStore struct {
	mu       sync.RWMutex
	sessions map[string]storedSession
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		sessions: make(map[string]storedSession),
	}
}

func (s *MemoryStore) Create(_ context.Context, username string) (string, error) {
	token := uuid.New().String()
	s.mu.Lock()
	s.sessions[token] = storedSession{
		Username:  username,
		ExpiresAt: time.Now().Add(sessionTTL),
	}
	s.mu.Unlock()
	return token, nil
}

func (s *MemoryStore) Validate(_ context.Context, token string) (string, error) {
	s.mu.RLock()
	sess, ok := s.sessions[token]
	s.mu.RUnlock()

	if !ok {
		return "", errors.NewAppError(errors.Unauthorized, "invalid or expired session")
	}
	if time.Now().After(sess.ExpiresAt) {
		s.delete(token)
		return "", errors.NewAppError(errors.Unauthorized, "invalid or expired session")
	}
	return sess.Username, nil
}

func (s *MemoryStore) Delete(_ context.Context, token string) error {
	s.delete(token)
	return nil
}

func (s *MemoryStore) delete(token string) {
	s.mu.Lock()
	delete(s.sessions, token)
	s.mu.Unlock()
}

func (s *MemoryStore) StartCleanupLoop(done <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.cleanExpired()
			case <-done:
				return
			}
		}
	}()
}

func (s *MemoryStore) cleanExpired() {
	now := time.Now()
	s.mu.Lock()
	for token, sess := range s.sessions {
		if now.After(sess.ExpiresAt) {
			delete(s.sessions, token)
		}
	}
	s.mu.Unlock()
}
