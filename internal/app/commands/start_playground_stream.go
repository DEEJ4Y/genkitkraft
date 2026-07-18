package commands

import (
	"context"
	"time"

	"github.com/rs/zerolog"

	"github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	chatprovider "github.com/DEEJ4Y/genkitkraft/internal/ports/chat_provider"
	playgroundrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/playground_repo"
	streamregistry "github.com/DEEJ4Y/genkitkraft/internal/ports/stream_registry"
)

// StreamGenerationTimeout bounds a detached generation goroutine so it can
// never run forever if a provider hangs. The SSE tail loop in the HTTP
// handler uses this same bound (plus a grace period) to decide when a
// "streaming" message with no recent chunk activity must be orphaned.
const StreamGenerationTimeout = 5 * time.Minute

type StartPlaygroundStreamParams struct {
	SessionID   string
	ChatRequest chatprovider.ChatRequest
}

type StartPlaygroundStreamResult struct {
	MessageID string
}

// StartPlaygroundStreamCommand creates the assistant message row and launches
// generation on a context detached from the caller's — a client disconnect
// must not abort the model call, only the HTTP handler tailing it. The
// returned MessageID is available immediately; the goroutine persists tokens
// and finalizes the message's status on its own.
type StartPlaygroundStreamCommand struct {
	repo         playgroundrepo.PlaygroundRepository
	chatProvider chatprovider.ChatProvider
	registry     streamregistry.Registry
	logger       zerolog.Logger
}

func NewStartPlaygroundStreamCommand(repo playgroundrepo.PlaygroundRepository, chatProvider chatprovider.ChatProvider, registry streamregistry.Registry, logger zerolog.Logger) *StartPlaygroundStreamCommand {
	return &StartPlaygroundStreamCommand{repo: repo, chatProvider: chatProvider, registry: registry, logger: logger}
}

func (c *StartPlaygroundStreamCommand) Execute(ctx context.Context, params StartPlaygroundStreamParams) (StartPlaygroundStreamResult, error) {
	if params.SessionID == "" {
		return StartPlaygroundStreamResult{}, errors.NewAppError(errors.InvalidInput, "session ID is required")
	}

	msg, err := c.repo.CreateStreamingMessage(ctx, params.SessionID)
	if err != nil {
		return StartPlaygroundStreamResult{}, err
	}

	genCtx, cancel := context.WithTimeout(context.Background(), StreamGenerationTimeout)
	c.registry.Register(msg.ID, cancel)

	go c.runGeneration(genCtx, cancel, msg.ID, params.ChatRequest)

	return StartPlaygroundStreamResult{MessageID: msg.ID}, nil
}

func (c *StartPlaygroundStreamCommand) runGeneration(ctx context.Context, cancel context.CancelFunc, messageID string, req chatprovider.ChatRequest) {
	defer cancel()
	defer c.registry.Unregister(messageID)

	// Persistence must survive both a timeout and an explicit cancel (the
	// "stop" button) so whatever was already generated is never lost — so
	// every repo write here uses a value-preserving, cancellation-free copy
	// of ctx rather than ctx itself.
	persistCtx := context.WithoutCancel(ctx)

	tokenCh, errCh := c.chatProvider.ChatStream(ctx, req)

	for token := range tokenCh {
		if _, err := c.repo.AppendMessageChunk(persistCtx, messageID, token); err != nil {
			c.logger.Error().Err(err).Str("message_id", messageID).Msg("appending playground stream chunk failed")
		}
	}

	if streamErr := <-errCh; streamErr != nil {
		c.logger.Warn().Err(streamErr).Str("message_id", messageID).Msg("playground stream generation failed")
		if err := c.repo.FailMessage(persistCtx, messageID); err != nil {
			c.logger.Error().Err(err).Str("message_id", messageID).Msg("marking playground stream message failed")
		}
		return
	}

	if err := c.repo.CompleteMessage(persistCtx, messageID); err != nil {
		c.logger.Error().Err(err).Str("message_id", messageID).Msg("marking playground stream message complete")
	}
}
