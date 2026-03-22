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

// Message represents a single message in a playground session.
type Message struct {
	ID        string
	SessionID string
	Role      string // "user" or "assistant"
	Content   string
	CreatedAt time.Time
}
