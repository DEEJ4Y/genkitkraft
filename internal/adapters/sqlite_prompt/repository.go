package sqliteprompt

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	apperrors "github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/prompt"
	promptrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/prompt_repo"
)

// Compile-time check that PromptRepository implements the port interface.
var _ promptrepo.PromptRepository = (*PromptRepository)(nil)

// PromptRepository implements promptrepo.PromptRepository using SQLite.
type PromptRepository struct {
	db *sql.DB
}

// NewPromptRepository creates a new SQLite-backed prompt repository.
func NewPromptRepository(db *sql.DB) *PromptRepository {
	return &PromptRepository{db: db}
}

func (r *PromptRepository) List(ctx context.Context, limit, offset int) ([]*prompt.Prompt, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, content, created_at, updated_at
		 FROM prompts ORDER BY created_at DESC LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, apperrors.NewAppErrorf(apperrors.Internal, "listing prompts: %v", err)
	}
	defer rows.Close()

	var prompts []*prompt.Prompt
	for rows.Next() {
		var p prompt.Prompt
		if err := rows.Scan(&p.ID, &p.Name, &p.Content, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, apperrors.NewAppErrorf(apperrors.Internal, "scanning prompt: %v", err)
		}
		prompts = append(prompts, &p)
	}
	return prompts, rows.Err()
}

func (r *PromptRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM prompts`).Scan(&count)
	if err != nil {
		return 0, apperrors.NewAppErrorf(apperrors.Internal, "counting prompts: %v", err)
	}
	return count, nil
}

func (r *PromptRepository) GetByID(ctx context.Context, id string) (*prompt.Prompt, error) {
	var p prompt.Prompt
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, content, created_at, updated_at
		 FROM prompts WHERE id = ?`, id).Scan(&p.ID, &p.Name, &p.Content, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, apperrors.NewAppError(apperrors.NotFound, "prompt not found")
	}
	if err != nil {
		return nil, apperrors.NewAppErrorf(apperrors.Internal, "getting prompt: %v", err)
	}
	return &p, nil
}

func (r *PromptRepository) Create(ctx context.Context, p *prompt.Prompt) error {
	p.ID = uuid.New().String()
	now := time.Now().UTC()
	p.CreatedAt = now
	p.UpdatedAt = now

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO prompts (id, name, content, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?)`,
		p.ID, p.Name, p.Content, p.CreatedAt, p.UpdatedAt)
	if err != nil {
		return apperrors.NewAppErrorf(apperrors.Internal, "creating prompt: %v", err)
	}
	return nil
}

func (r *PromptRepository) Update(ctx context.Context, p *prompt.Prompt) error {
	p.UpdatedAt = time.Now().UTC()

	result, err := r.db.ExecContext(ctx,
		`UPDATE prompts SET name = ?, content = ?, updated_at = ? WHERE id = ?`,
		p.Name, p.Content, p.UpdatedAt, p.ID)
	if err != nil {
		return apperrors.NewAppErrorf(apperrors.Internal, "updating prompt: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return apperrors.NewAppErrorf(apperrors.Internal, "checking update result: %v", err)
	}
	if rows == 0 {
		return apperrors.NewAppError(apperrors.NotFound, "prompt not found")
	}
	return nil
}

func (r *PromptRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM prompts WHERE id = ?`, id)
	if err != nil {
		return apperrors.NewAppErrorf(apperrors.Internal, "deleting prompt: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return apperrors.NewAppErrorf(apperrors.Internal, "checking delete result: %v", err)
	}
	if rows == 0 {
		return apperrors.NewAppError(apperrors.NotFound, "prompt not found")
	}
	return nil
}
