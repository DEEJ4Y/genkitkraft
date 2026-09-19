package playground

import "time"

// Session represents a playground chat session for testing an agent.
type Session struct {
	ID        string
	AgentID   string
	Title     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Message status values. Existing rows (and any row created via CreateMessage) are
// implicitly "complete" — only streaming assistant replies pass through "streaming"
// and, on failure, "error".
const (
	MessageStatusStreaming = "streaming"
	MessageStatusComplete  = "complete"
	MessageStatusError     = "error"
)

// Message represents a single message in a playground session.
type Message struct {
	ID        string
	SessionID string
	Role      string // "user" or "assistant"
	Content   string
	Status    string // "streaming", "complete", or "error"
	CreatedAt time.Time
}

// MessageChunk is one persisted, sequence-numbered piece of a streaming
// assistant reply. Seq maps 1:1 to the SSE `id:` field for that token.
type MessageChunk struct {
	Seq     int
	Content string
}
