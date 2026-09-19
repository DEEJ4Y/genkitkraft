package queries

import (
	"context"

	apperrors "github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/playground"
	playgroundrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/playground_repo"
)

type GetPlaygroundStreamChunksParams struct {
	SessionID string
	// MessageID pins the tail to a specific message (the initial connection
	// already knows it from StartStream's result). Leave nil to resolve the
	// session's latest message instead — used by reconnect, which only knows
	// the session ID.
	MessageID *string
	SinceSeq  int
}

type GetPlaygroundStreamChunksResult struct {
	MessageID string
	Status    string
	Chunks    []playground.MessageChunk
}

// GetPlaygroundStreamChunksQuery serves both the initial SSE connection and a
// later reconnect: the only difference is SinceSeq (0 vs. Last-Event-ID) and
// whether MessageID is already known.
type GetPlaygroundStreamChunksQuery struct {
	repo playgroundrepo.PlaygroundRepository
}

func NewGetPlaygroundStreamChunksQuery(repo playgroundrepo.PlaygroundRepository) *GetPlaygroundStreamChunksQuery {
	return &GetPlaygroundStreamChunksQuery{repo: repo}
}

func (q *GetPlaygroundStreamChunksQuery) Execute(ctx context.Context, params GetPlaygroundStreamChunksParams) (GetPlaygroundStreamChunksResult, error) {
	if params.SessionID == "" {
		return GetPlaygroundStreamChunksResult{}, apperrors.NewAppError(apperrors.InvalidInput, "session ID is required")
	}

	messageID := ""
	if params.MessageID != nil {
		messageID = *params.MessageID
	}

	var msg *playground.Message
	var err error
	if messageID != "" {
		msg, err = q.repo.GetMessage(ctx, messageID)
	} else {
		msg, err = q.repo.GetLatestMessageBySession(ctx, params.SessionID)
	}
	if err != nil {
		return GetPlaygroundStreamChunksResult{}, err
	}
	if msg.SessionID != params.SessionID {
		return GetPlaygroundStreamChunksResult{}, apperrors.NewAppError(apperrors.NotFound, "playground message not found")
	}

	chunks, err := q.repo.GetMessageChunksSince(ctx, msg.ID, params.SinceSeq)
	if err != nil {
		return GetPlaygroundStreamChunksResult{}, err
	}

	return GetPlaygroundStreamChunksResult{MessageID: msg.ID, Status: msg.Status, Chunks: chunks}, nil
}
