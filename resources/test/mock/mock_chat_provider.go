package mock

import (
	"context"
	"time"

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
	// StreamDelay, if set, is waited between sending each token — long enough
	// for a test to cancel the context or simulate a disconnect mid-stream.
	// Zero (the default) sends every token immediately, matching the original
	// behavior.
	StreamDelay time.Duration
	// LastRequest captures the most recent ChatRequest for assertions.
	LastRequest chatprovider.ChatRequest
}

func (m *ChatProvider) Chat(_ context.Context, req chatprovider.ChatRequest) (string, error) {
	m.LastRequest = req
	return m.ChatResponse, m.ChatError
}

// ChatStream respects ctx cancellation, unlike a naive mock that would send
// every token regardless — callers that decouple generation from a request
// context (see StartPlaygroundStreamCommand) rely on ctx.Done() to actually
// stop the goroutine, and tests need that to be real to exercise cancel and
// mid-stream-disconnect scenarios.
func (m *ChatProvider) ChatStream(ctx context.Context, req chatprovider.ChatRequest) (<-chan string, <-chan error) {
	m.LastRequest = req

	tokenCh := make(chan string, len(m.StreamTokens))
	errCh := make(chan error, 1)

	go func() {
		defer close(tokenCh)
		defer close(errCh)

		for _, t := range m.StreamTokens {
			select {
			case tokenCh <- t:
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			}

			if m.StreamDelay > 0 {
				select {
				case <-time.After(m.StreamDelay):
				case <-ctx.Done():
					errCh <- ctx.Err()
					return
				}
			}
		}
		if m.StreamError != nil {
			errCh <- m.StreamError
		}
	}()

	return tokenCh, errCh
}
