package queries

import (
	"context"

	apperrors "github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/playground"
	playgroundrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/playground_repo"
)

type ListPlaygroundMessagesParams struct {
	SessionID string
	AgentID   string
}

type ListPlaygroundMessagesResult struct {
	Messages []*playground.Message
}

type ListPlaygroundMessagesQuery struct {
	repo playgroundrepo.PlaygroundRepository
}

func NewListPlaygroundMessagesQuery(repo playgroundrepo.PlaygroundRepository) *ListPlaygroundMessagesQuery {
	return &ListPlaygroundMessagesQuery{repo: repo}
}

func (q *ListPlaygroundMessagesQuery) Execute(ctx context.Context, params ListPlaygroundMessagesParams) (ListPlaygroundMessagesResult, error) {
	session, err := q.repo.GetSession(ctx, params.SessionID)
	if err != nil {
		return ListPlaygroundMessagesResult{}, err
	}
	if session.AgentID != params.AgentID {
		return ListPlaygroundMessagesResult{}, apperrors.NewAppError(apperrors.NotFound, "playground session not found")
	}
	messages, err := q.repo.ListMessagesBySession(ctx, params.SessionID)
	if err != nil {
		return ListPlaygroundMessagesResult{}, err
	}
	return ListPlaygroundMessagesResult{Messages: messages}, nil
}
