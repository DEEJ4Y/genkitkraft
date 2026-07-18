package commands

import (
	"context"

	apperrors "github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	playgroundrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/playground_repo"
	streamregistry "github.com/DEEJ4Y/genkitkraft/internal/ports/stream_registry"
)

type CancelPlaygroundStreamParams struct {
	SessionID string
}

// CancelPlaygroundStreamCommand stops an in-flight stream for a session, if
// one is running on this instance. It is intentionally a no-op — not an
// error — when there is nothing to cancel: the generation may have already
// finished, or be running on a different instance, and "stop" arriving too
// late is a normal race, not a failure.
type CancelPlaygroundStreamCommand struct {
	repo     playgroundrepo.PlaygroundRepository
	registry streamregistry.Registry
}

func NewCancelPlaygroundStreamCommand(repo playgroundrepo.PlaygroundRepository, registry streamregistry.Registry) *CancelPlaygroundStreamCommand {
	return &CancelPlaygroundStreamCommand{repo: repo, registry: registry}
}

func (c *CancelPlaygroundStreamCommand) Execute(ctx context.Context, params CancelPlaygroundStreamParams) error {
	if params.SessionID == "" {
		return apperrors.NewAppError(apperrors.InvalidInput, "session ID is required")
	}

	msg, err := c.repo.GetLatestMessageBySession(ctx, params.SessionID)
	if err != nil {
		if appErr, ok := apperrors.IsAppError(err); ok && appErr.Code() == apperrors.NotFound {
			return nil
		}
		return err
	}

	c.registry.Cancel(msg.ID)
	return nil
}
