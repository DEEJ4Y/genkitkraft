package sqliteagent

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	apperrors "github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/agent"
	agentrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/agent_repo"
)

// Compile-time check that AgentRepository implements the port interface.
var _ agentrepo.AgentRepository = (*AgentRepository)(nil)

// AgentRepository implements agentrepo.AgentRepository using SQLite.
type AgentRepository struct {
	db *sql.DB
}

// NewAgentRepository creates a new SQLite-backed agent repository.
func NewAgentRepository(db *sql.DB) *AgentRepository {
	return &AgentRepository{db: db}
}

const listQuery = `
SELECT a.id, a.name, a.provider_id, a.model_id, a.system_prompt_id,
       a.temperature, a.top_p, a.top_k, a.created_at, a.updated_at,
       COALESCE(p.name, '') AS provider_name,
       COALESCE(p.provider_type, '') AS provider_type,
       COALESCE(pr.name, '') AS system_prompt_name
FROM agents a
LEFT JOIN providers p ON a.provider_id = p.id
LEFT JOIN prompts pr ON a.system_prompt_id = pr.id
ORDER BY a.created_at DESC LIMIT ? OFFSET ?`

const getByIDQuery = `
SELECT a.id, a.name, a.provider_id, a.model_id, a.system_prompt_id,
       a.temperature, a.top_p, a.top_k, a.created_at, a.updated_at,
       COALESCE(p.name, '') AS provider_name,
       COALESCE(p.provider_type, '') AS provider_type,
       COALESCE(pr.name, '') AS system_prompt_name
FROM agents a
LEFT JOIN providers p ON a.provider_id = p.id
LEFT JOIN prompts pr ON a.system_prompt_id = pr.id
WHERE a.id = ?`

func (r *AgentRepository) List(ctx context.Context, limit, offset int) ([]*agent.Agent, error) {
	rows, err := r.db.QueryContext(ctx, listQuery, limit, offset)
	if err != nil {
		return nil, apperrors.NewAppErrorf(apperrors.Internal, "listing agents: %v", err)
	}
	defer rows.Close()

	var agents []*agent.Agent
	for rows.Next() {
		a, err := scanAgent(rows)
		if err != nil {
			return nil, err
		}
		agents = append(agents, a)
	}
	return agents, rows.Err()
}

func (r *AgentRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM agents`).Scan(&count)
	if err != nil {
		return 0, apperrors.NewAppErrorf(apperrors.Internal, "counting agents: %v", err)
	}
	return count, nil
}

func (r *AgentRepository) GetByID(ctx context.Context, id string) (*agent.Agent, error) {
	row := r.db.QueryRowContext(ctx, getByIDQuery, id)
	a, err := scanAgentRow(row)
	if err == sql.ErrNoRows {
		return nil, apperrors.NewAppError(apperrors.NotFound, "agent not found")
	}
	if err != nil {
		return nil, apperrors.NewAppErrorf(apperrors.Internal, "getting agent: %v", err)
	}
	return a, nil
}

func (r *AgentRepository) Create(ctx context.Context, a *agent.Agent) error {
	a.ID = uuid.New().String()
	now := time.Now().UTC()
	a.CreatedAt = now
	a.UpdatedAt = now

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO agents (id, name, provider_id, model_id, system_prompt_id, temperature, top_p, top_k, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		a.ID, a.Name, a.ProviderID, a.ModelID, nullString(a.SystemPromptID),
		a.Temperature, a.TopP, a.TopK, a.CreatedAt, a.UpdatedAt)
	if err != nil {
		return apperrors.NewAppErrorf(apperrors.Internal, "creating agent: %v", err)
	}
	return nil
}

func (r *AgentRepository) Update(ctx context.Context, a *agent.Agent) error {
	a.UpdatedAt = time.Now().UTC()

	result, err := r.db.ExecContext(ctx,
		`UPDATE agents SET name = ?, provider_id = ?, model_id = ?, system_prompt_id = ?,
		 temperature = ?, top_p = ?, top_k = ?, updated_at = ? WHERE id = ?`,
		a.Name, a.ProviderID, a.ModelID, nullString(a.SystemPromptID),
		a.Temperature, a.TopP, a.TopK, a.UpdatedAt, a.ID)
	if err != nil {
		return apperrors.NewAppErrorf(apperrors.Internal, "updating agent: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return apperrors.NewAppErrorf(apperrors.Internal, "checking update result: %v", err)
	}
	if rows == 0 {
		return apperrors.NewAppError(apperrors.NotFound, "agent not found")
	}
	return nil
}

func (r *AgentRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM agents WHERE id = ?`, id)
	if err != nil {
		return apperrors.NewAppErrorf(apperrors.Internal, "deleting agent: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return apperrors.NewAppErrorf(apperrors.Internal, "checking delete result: %v", err)
	}
	if rows == 0 {
		return apperrors.NewAppError(apperrors.NotFound, "agent not found")
	}
	return nil
}

func scanAgent(rows *sql.Rows) (*agent.Agent, error) {
	var a agent.Agent
	var systemPromptID sql.NullString
	if err := rows.Scan(
		&a.ID, &a.Name, &a.ProviderID, &a.ModelID, &systemPromptID,
		&a.Temperature, &a.TopP, &a.TopK, &a.CreatedAt, &a.UpdatedAt,
		&a.ProviderName, &a.ProviderType, &a.SystemPromptName,
	); err != nil {
		return nil, apperrors.NewAppErrorf(apperrors.Internal, "scanning agent: %v", err)
	}
	if systemPromptID.Valid {
		a.SystemPromptID = systemPromptID.String
	}
	return &a, nil
}

func scanAgentRow(row *sql.Row) (*agent.Agent, error) {
	var a agent.Agent
	var systemPromptID sql.NullString
	if err := row.Scan(
		&a.ID, &a.Name, &a.ProviderID, &a.ModelID, &systemPromptID,
		&a.Temperature, &a.TopP, &a.TopK, &a.CreatedAt, &a.UpdatedAt,
		&a.ProviderName, &a.ProviderType, &a.SystemPromptName,
	); err != nil {
		return nil, err // caller handles sql.ErrNoRows
	}
	if systemPromptID.Valid {
		a.SystemPromptID = systemPromptID.String
	}
	return &a, nil
}

// nullString converts a string to sql.NullString. Empty string maps to NULL.
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
