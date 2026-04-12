package commands

import (
	"context"

	"github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/playground"
	playgroundrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/playground_repo"
)

type SavePlaygroundMessageParams struct {
	SessionID string
	Role      string
	Content   string
}

type SavePlaygroundMessageResult struct {
	Message *playground.Message
}

type SavePlaygroundMessageCommand struct {
	repo playgroundrepo.PlaygroundRepository
}

func NewSavePlaygroundMessageCommand(repo playgroundrepo.PlaygroundRepository) *SavePlaygroundMessageCommand {
	return &SavePlaygroundMessageCommand{repo: repo}
}

func (c *SavePlaygroundMessageCommand) Execute(ctx context.Context, params SavePlaygroundMessageParams) (SavePlaygroundMessageResult, error) {
	if params.SessionID == "" {
		return SavePlaygroundMessageResult{}, errors.NewAppError(errors.InvalidInput, "session ID is required")
	}
	if params.Content == "" {
		return SavePlaygroundMessageResult{}, errors.NewAppError(errors.InvalidInput, "content is required")
	}
	if params.Role != "user" && params.Role != "assistant" {
		return SavePlaygroundMessageResult{}, errors.NewAppError(errors.InvalidInput, "role must be 'user' or 'assistant'")
	}

	m := &playground.Message{
		SessionID: params.SessionID,
		Role:      params.Role,
		Content:   params.Content,
	}

	if err := c.repo.CreateMessage(ctx, m); err != nil {
		return SavePlaygroundMessageResult{}, err
	}

	// Auto-title: update session title from first user message
	if params.Role == "user" {
		title := params.Content
		if len(title) > 50 {
			title = title[:50] + "..."
		}
		// Ignore error — auto-titling is best-effort
		_ = c.repo.UpdateSessionTitle(ctx, params.SessionID, title)
	}

	return SavePlaygroundMessageResult{Message: m}, nil
}
