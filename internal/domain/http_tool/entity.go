package httptool

import "time"

// HttpToolHeader represents a key-value pair for an HTTP header.
type HttpToolHeader struct {
	Name  string
	Value string
}

// HttpTool represents an HTTP tool that can be attached to agents for LLM function calling.
type HttpTool struct {
	ID           string
	Name         string
	Description  string
	Method       string
	URL          string
	Headers      []HttpToolHeader
	BodyTemplate string
	InputSchema  string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
