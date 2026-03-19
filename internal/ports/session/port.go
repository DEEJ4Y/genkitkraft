package session

import "context"

// Store defines the contract for session persistence.
type Store interface {
	Create(ctx context.Context, username string) (token string, err error)
	Validate(ctx context.Context, token string) (username string, err error)
	Delete(ctx context.Context, token string) error
	StartCleanupLoop(done <-chan struct{})
}
