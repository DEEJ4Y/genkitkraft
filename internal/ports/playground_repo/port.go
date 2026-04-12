package playgroundrepo

import (
	"context"

	"github.com/DEEJ4Y/genkitkraft/internal/domain/playground"
)

// PlaygroundRepository defines the contract for playground session and message persistence.
type PlaygroundRepository interface {
	CreateSession(ctx context.Context, s *playground.Session) error
	GetSession(ctx context.Context, id string) (*playground.Session, error)
	ListSessionsByAgent(ctx context.Context, agentID string) ([]*playground.Session, error)
	DeleteSession(ctx context.Context, id string) error
	UpdateSessionTitle(ctx context.Context, id, title string) error

	CreateMessage(ctx context.Context, m *playground.Message) error
	ListMessagesBySession(ctx context.Context, sessionID string) ([]*playground.Message, error)
}
