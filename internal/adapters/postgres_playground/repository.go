package postgresplayground

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	apperrors "github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/playground"
	playgroundrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/playground_repo"
)

// Compile-time check that PlaygroundRepository implements the port interface.
var _ playgroundrepo.PlaygroundRepository = (*PlaygroundRepository)(nil)

// PlaygroundRepository implements playgroundrepo.PlaygroundRepository using PostgreSQL.
type PlaygroundRepository struct {
	db *sql.DB
}

// NewPlaygroundRepository creates a new PostgreSQL-backed playground repository.
func NewPlaygroundRepository(db *sql.DB) *PlaygroundRepository {
	return &PlaygroundRepository{db: db}
}

func (r *PlaygroundRepository) CreateSession(ctx context.Context, s *playground.Session) error {
	s.ID = uuid.New().String()
	now := time.Now().UTC()
	s.CreatedAt = now
	s.UpdatedAt = now

	if s.Title == "" {
		s.Title = "New Session"
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO playground_sessions (id, agent_id, title, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		s.ID, s.AgentID, s.Title, s.CreatedAt, s.UpdatedAt)
	if err != nil {
		return apperrors.NewAppErrorf(apperrors.Internal, "creating playground session: %v", err)
	}
	return nil
}

func (r *PlaygroundRepository) GetSession(ctx context.Context, id string) (*playground.Session, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, agent_id, title, created_at, updated_at FROM playground_sessions WHERE id = $1`, id)

	var s playground.Session
	err := row.Scan(&s.ID, &s.AgentID, &s.Title, &s.CreatedAt, &s.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, apperrors.NewAppError(apperrors.NotFound, "playground session not found")
	}
	if err != nil {
		return nil, apperrors.NewAppErrorf(apperrors.Internal, "getting playground session: %v", err)
	}
	return &s, nil
}

func (r *PlaygroundRepository) ListSessionsByAgent(ctx context.Context, agentID string) ([]*playground.Session, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, agent_id, title, created_at, updated_at
		 FROM playground_sessions WHERE agent_id = $1 ORDER BY updated_at DESC`, agentID)
	if err != nil {
		return nil, apperrors.NewAppErrorf(apperrors.Internal, "listing playground sessions: %v", err)
	}
	defer rows.Close()

	var sessions []*playground.Session
	for rows.Next() {
		var s playground.Session
		if err := rows.Scan(&s.ID, &s.AgentID, &s.Title, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, apperrors.NewAppErrorf(apperrors.Internal, "scanning playground session: %v", err)
		}
		sessions = append(sessions, &s)
	}
	return sessions, rows.Err()
}

func (r *PlaygroundRepository) DeleteSession(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM playground_sessions WHERE id = $1`, id)
	if err != nil {
		return apperrors.NewAppErrorf(apperrors.Internal, "deleting playground session: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return apperrors.NewAppErrorf(apperrors.Internal, "checking delete result: %v", err)
	}
	if rows == 0 {
		return apperrors.NewAppError(apperrors.NotFound, "playground session not found")
	}
	return nil
}

func (r *PlaygroundRepository) UpdateSessionTitle(ctx context.Context, id, title string) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE playground_sessions SET title = $1, updated_at = $2 WHERE id = $3`,
		title, time.Now().UTC(), id)
	if err != nil {
		return apperrors.NewAppErrorf(apperrors.Internal, "updating session title: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return apperrors.NewAppErrorf(apperrors.Internal, "checking update result: %v", err)
	}
	if rows == 0 {
		return apperrors.NewAppError(apperrors.NotFound, "playground session not found")
	}
	return nil
}

func (r *PlaygroundRepository) CreateMessage(ctx context.Context, m *playground.Message) error {
	m.ID = uuid.New().String()
	m.CreatedAt = time.Now().UTC()
	m.Status = playground.MessageStatusComplete

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO playground_messages (id, session_id, role, content, status, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		m.ID, m.SessionID, m.Role, m.Content, m.Status, m.CreatedAt)
	if err != nil {
		return apperrors.NewAppErrorf(apperrors.Internal, "creating playground message: %v", err)
	}
	return nil
}

func (r *PlaygroundRepository) ListMessagesBySession(ctx context.Context, sessionID string) ([]*playground.Message, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, session_id, role, content, status, created_at
		 FROM playground_messages WHERE session_id = $1 ORDER BY created_at ASC`, sessionID)
	if err != nil {
		return nil, apperrors.NewAppErrorf(apperrors.Internal, "listing playground messages: %v", err)
	}
	defer rows.Close()

	var messages []*playground.Message
	for rows.Next() {
		var m playground.Message
		if err := rows.Scan(&m.ID, &m.SessionID, &m.Role, &m.Content, &m.Status, &m.CreatedAt); err != nil {
			return nil, apperrors.NewAppErrorf(apperrors.Internal, "scanning playground message: %v", err)
		}
		messages = append(messages, &m)
	}
	return messages, rows.Err()
}

func (r *PlaygroundRepository) CreateStreamingMessage(ctx context.Context, sessionID string) (*playground.Message, error) {
	m := &playground.Message{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Role:      "assistant",
		Content:   "",
		Status:    playground.MessageStatusStreaming,
		CreatedAt: time.Now().UTC(),
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO playground_messages (id, session_id, role, content, status, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		m.ID, m.SessionID, m.Role, m.Content, m.Status, m.CreatedAt)
	if err != nil {
		return nil, apperrors.NewAppErrorf(apperrors.Internal, "creating streaming message: %v", err)
	}
	return m, nil
}

// AppendMessageChunk is only ever called sequentially for a given messageID
// (from the single generation goroutine that owns it), so the seq
// read-then-insert below needs no extra locking beyond the transaction.
func (r *PlaygroundRepository) AppendMessageChunk(ctx context.Context, messageID string, chunk string) (int, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, apperrors.NewAppErrorf(apperrors.Internal, "starting transaction: %v", err)
	}
	defer tx.Rollback()

	var seq int
	if err := tx.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(seq), 0) + 1 FROM playground_message_chunks WHERE message_id = $1`, messageID,
	).Scan(&seq); err != nil {
		return 0, apperrors.NewAppErrorf(apperrors.Internal, "computing next chunk seq: %v", err)
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO playground_message_chunks (id, message_id, seq, content, created_at) VALUES ($1, $2, $3, $4, $5)`,
		uuid.New().String(), messageID, seq, chunk, time.Now().UTC())
	if err != nil {
		return 0, apperrors.NewAppErrorf(apperrors.Internal, "inserting message chunk: %v", err)
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE playground_messages SET content = content || $1 WHERE id = $2`, chunk, messageID,
	); err != nil {
		return 0, apperrors.NewAppErrorf(apperrors.Internal, "appending message content: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, apperrors.NewAppErrorf(apperrors.Internal, "committing message chunk: %v", err)
	}
	return seq, nil
}

func (r *PlaygroundRepository) GetMessageChunksSince(ctx context.Context, messageID string, sinceSeq int) ([]playground.MessageChunk, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT seq, content FROM playground_message_chunks WHERE message_id = $1 AND seq > $2 ORDER BY seq ASC`,
		messageID, sinceSeq)
	if err != nil {
		return nil, apperrors.NewAppErrorf(apperrors.Internal, "listing message chunks: %v", err)
	}
	defer rows.Close()

	var chunks []playground.MessageChunk
	for rows.Next() {
		var c playground.MessageChunk
		if err := rows.Scan(&c.Seq, &c.Content); err != nil {
			return nil, apperrors.NewAppErrorf(apperrors.Internal, "scanning message chunk: %v", err)
		}
		chunks = append(chunks, c)
	}
	return chunks, rows.Err()
}

func (r *PlaygroundRepository) CompleteMessage(ctx context.Context, messageID string) error {
	return r.setStreamingMessageStatus(ctx, messageID, playground.MessageStatusComplete)
}

func (r *PlaygroundRepository) FailMessage(ctx context.Context, messageID string) error {
	return r.setStreamingMessageStatus(ctx, messageID, playground.MessageStatusError)
}

func (r *PlaygroundRepository) setStreamingMessageStatus(ctx context.Context, messageID, status string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE playground_messages SET status = $1 WHERE id = $2 AND status = $3`,
		status, messageID, playground.MessageStatusStreaming)
	if err != nil {
		return apperrors.NewAppErrorf(apperrors.Internal, "updating message status: %v", err)
	}
	return nil
}

func (r *PlaygroundRepository) GetMessage(ctx context.Context, id string) (*playground.Message, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, session_id, role, content, status, created_at FROM playground_messages WHERE id = $1`, id)

	var m playground.Message
	err := row.Scan(&m.ID, &m.SessionID, &m.Role, &m.Content, &m.Status, &m.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, apperrors.NewAppError(apperrors.NotFound, "playground message not found")
	}
	if err != nil {
		return nil, apperrors.NewAppErrorf(apperrors.Internal, "getting playground message: %v", err)
	}
	return &m, nil
}

func (r *PlaygroundRepository) GetLatestMessageBySession(ctx context.Context, sessionID string) (*playground.Message, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, session_id, role, content, status, created_at
		 FROM playground_messages WHERE session_id = $1 ORDER BY created_at DESC LIMIT 1`, sessionID)

	var m playground.Message
	err := row.Scan(&m.ID, &m.SessionID, &m.Role, &m.Content, &m.Status, &m.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, apperrors.NewAppError(apperrors.NotFound, "playground message not found")
	}
	if err != nil {
		return nil, apperrors.NewAppErrorf(apperrors.Internal, "getting latest playground message: %v", err)
	}
	return &m, nil
}
