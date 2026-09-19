package commands

import (
	"context"

	apperrors "github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	playgroundrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/playground_repo"
)

type FailPlaygroundStreamParams struct {
	MessageID string
}

// FailPlaygroundStreamCommand marks a message as errored. It exists for the
// HTTP tail loop's orphan-detection path: a message stuck in "streaming" with
// no chunk activity past the generation's own timeout window means the
// goroutine that would have finalized it is gone (e.g. its instance crashed),
// so the tailer finalizes it instead.
type FailPlaygroundStreamCommand struct {
	repo playgroundrepo.PlaygroundRepository
}

func NewFailPlaygroundStreamCommand(repo playgroundrepo.PlaygroundRepository) *FailPlaygroundStreamCommand {
	return &FailPlaygroundStreamCommand{repo: repo}
}

func (c *FailPlaygroundStreamCommand) Execute(ctx context.Context, params FailPlaygroundStreamParams) error {
	if params.MessageID == "" {
		return apperrors.NewAppError(apperrors.InvalidInput, "message ID is required")
	}
	return c.repo.FailMessage(ctx, params.MessageID)
}
