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

	// CreateStreamingMessage inserts a new assistant message row with empty
	// content and status "streaming", to be filled in via AppendMessageChunk
	// as generation proceeds.
	CreateStreamingMessage(ctx context.Context, sessionID string) (*playground.Message, error)
	// AppendMessageChunk assigns the next sequence number for messageID,
	// persists chunk as its own row, and appends chunk to the message's
	// content — all as a single atomic write. The returned seq is the SSE
	// `id:` value for this chunk.
	AppendMessageChunk(ctx context.Context, messageID string, chunk string) (seq int, err error)
	// GetMessageChunksSince returns chunks with seq > sinceSeq for messageID,
	// ordered by seq ascending.
	GetMessageChunksSince(ctx context.Context, messageID string, sinceSeq int) ([]playground.MessageChunk, error)
	// CompleteMessage marks messageID as complete. No-op (not an error) if the
	// message is not currently "streaming" — guards against a race with a
	// concurrent FailMessage.
	CompleteMessage(ctx context.Context, messageID string) error
	// FailMessage marks messageID as errored, preserving whatever content was
	// already appended. No-op (not an error) if the message is not currently
	// "streaming".
	FailMessage(ctx context.Context, messageID string) error
	// GetMessage fetches a single message by ID.
	GetMessage(ctx context.Context, id string) (*playground.Message, error)
	// GetLatestMessageBySession returns the most recently created message in
	// the session, used to resolve which message a reconnect should tail
	// without the client needing to know its ID.
	GetLatestMessageBySession(ctx context.Context, sessionID string) (*playground.Message, error)
}
