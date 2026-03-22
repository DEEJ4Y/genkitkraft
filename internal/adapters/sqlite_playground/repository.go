package sqliteplayground

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

// PlaygroundRepository implements playgroundrepo.PlaygroundRepository using SQLite.
type PlaygroundRepository struct {
	db *sql.DB
}

// NewPlaygroundRepository creates a new SQLite-backed playground repository.
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
		 VALUES (?, ?, ?, ?, ?)`,
		s.ID, s.AgentID, s.Title, s.CreatedAt, s.UpdatedAt)
	if err != nil {
		return apperrors.NewAppErrorf(apperrors.Internal, "creating playground session: %v", err)
	}
	return nil
}

func (r *PlaygroundRepository) GetSession(ctx context.Context, id string) (*playground.Session, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, agent_id, title, created_at, updated_at FROM playground_sessions WHERE id = ?`, id)

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
		 FROM playground_sessions WHERE agent_id = ? ORDER BY updated_at DESC`, agentID)
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
	result, err := r.db.ExecContext(ctx, `DELETE FROM playground_sessions WHERE id = ?`, id)
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
		`UPDATE playground_sessions SET title = ?, updated_at = ? WHERE id = ?`,
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

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO playground_messages (id, session_id, role, content, created_at)
		 VALUES (?, ?, ?, ?, ?)`,
		m.ID, m.SessionID, m.Role, m.Content, m.CreatedAt)
	if err != nil {
		return apperrors.NewAppErrorf(apperrors.Internal, "creating playground message: %v", err)
	}
	return nil
}

func (r *PlaygroundRepository) ListMessagesBySession(ctx context.Context, sessionID string) ([]*playground.Message, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, session_id, role, content, created_at
		 FROM playground_messages WHERE session_id = ? ORDER BY created_at ASC`, sessionID)
	if err != nil {
		return nil, apperrors.NewAppErrorf(apperrors.Internal, "listing playground messages: %v", err)
	}
	defer rows.Close()

	var messages []*playground.Message
	for rows.Next() {
		var m playground.Message
		if err := rows.Scan(&m.ID, &m.SessionID, &m.Role, &m.Content, &m.CreatedAt); err != nil {
			return nil, apperrors.NewAppErrorf(apperrors.Internal, "scanning playground message: %v", err)
		}
		messages = append(messages, &m)
	}
	return messages, rows.Err()
}
