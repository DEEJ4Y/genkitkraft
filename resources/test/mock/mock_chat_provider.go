package mock

import (
	"context"

	chatprovider "github.com/DEEJ4Y/genkitkraft/internal/ports/chat_provider"
)

// Compile-time check.
var _ chatprovider.ChatProvider = (*ChatProvider)(nil)

// ChatProvider is a mock implementation of the ChatProvider port for testing.
type ChatProvider struct {
	// ChatResponse is returned by Chat(). Set before calling.
	ChatResponse string
	// ChatError is returned by Chat() if non-nil.
	ChatError error
	// StreamTokens are sent one at a time by ChatStream().
	StreamTokens []string
	// StreamError is sent on the error channel if non-nil.
	StreamError error
	// LastRequest captures the most recent ChatRequest for assertions.
	LastRequest chatprovider.ChatRequest
}

func (m *ChatProvider) Chat(_ context.Context, req chatprovider.ChatRequest) (string, error) {
	m.LastRequest = req
	return m.ChatResponse, m.ChatError
}

func (m *ChatProvider) ChatStream(_ context.Context, req chatprovider.ChatRequest) (<-chan string, <-chan error) {
	m.LastRequest = req

	tokenCh := make(chan string, len(m.StreamTokens))
	errCh := make(chan error, 1)

	go func() {
		for _, t := range m.StreamTokens {
			tokenCh <- t
		}
		close(tokenCh)
		errCh <- m.StreamError
	}()

	return tokenCh, errCh
}
