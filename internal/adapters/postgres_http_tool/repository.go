package postgreshttptool

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	apperrors "github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	httptool "github.com/DEEJ4Y/genkitkraft/internal/domain/http_tool"
	httptoolrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/http_tool_repo"
)

// Compile-time check that HttpToolRepository implements the port interface.
var _ httptoolrepo.HttpToolRepository = (*HttpToolRepository)(nil)

// HttpToolRepository implements httptoolrepo.HttpToolRepository using PostgreSQL.
type HttpToolRepository struct {
	db *sql.DB
}

// NewHttpToolRepository creates a new PostgreSQL-backed HTTP tool repository.
func NewHttpToolRepository(db *sql.DB) *HttpToolRepository {
	return &HttpToolRepository{db: db}
}

func (r *HttpToolRepository) List(ctx context.Context, limit, offset int) ([]*httptool.HttpTool, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, description, method, url, headers, body_template, input_schema, created_at, updated_at
		 FROM http_tools ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, apperrors.NewAppErrorf(apperrors.Internal, "listing http tools: %v", err)
	}
	defer rows.Close()

	var tools []*httptool.HttpTool
	for rows.Next() {
		t, err := scanHttpTool(rows)
		if err != nil {
			return nil, err
		}
		tools = append(tools, t)
	}
	return tools, rows.Err()
}

func (r *HttpToolRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM http_tools`).Scan(&count)
	if err != nil {
		return 0, apperrors.NewAppErrorf(apperrors.Internal, "counting http tools: %v", err)
	}
	return count, nil
}

func (r *HttpToolRepository) GetByID(ctx context.Context, id string) (*httptool.HttpTool, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, description, method, url, headers, body_template, input_schema, created_at, updated_at
		 FROM http_tools WHERE id = $1`, id)

	t, err := scanHttpToolRow(row)
	if err == sql.ErrNoRows {
		return nil, apperrors.NewAppError(apperrors.NotFound, "http tool not found")
	}
	if err != nil {
		return nil, apperrors.NewAppErrorf(apperrors.Internal, "getting http tool: %v", err)
	}
	return t, nil
}

func (r *HttpToolRepository) Create(ctx context.Context, t *httptool.HttpTool) error {
	t.ID = uuid.New().String()
	now := time.Now().UTC()
	t.CreatedAt = now
	t.UpdatedAt = now

	headersJSON, err := json.Marshal(t.Headers)
	if err != nil {
		return apperrors.NewAppErrorf(apperrors.Internal, "marshaling headers: %v", err)
	}

	_, err = r.db.ExecContext(ctx,
		`INSERT INTO http_tools (id, name, description, method, url, headers, body_template, input_schema, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		t.ID, t.Name, t.Description, t.Method, t.URL, string(headersJSON), t.BodyTemplate, t.InputSchema, t.CreatedAt, t.UpdatedAt)
	if err != nil {
		return apperrors.NewAppErrorf(apperrors.Internal, "creating http tool: %v", err)
	}
	return nil
}

func (r *HttpToolRepository) Update(ctx context.Context, t *httptool.HttpTool) error {
	t.UpdatedAt = time.Now().UTC()

	headersJSON, err := json.Marshal(t.Headers)
	if err != nil {
		return apperrors.NewAppErrorf(apperrors.Internal, "marshaling headers: %v", err)
	}

	result, err := r.db.ExecContext(ctx,
		`UPDATE http_tools SET name = $1, description = $2, method = $3, url = $4, headers = $5, body_template = $6, input_schema = $7, updated_at = $8 WHERE id = $9`,
		t.Name, t.Description, t.Method, t.URL, string(headersJSON), t.BodyTemplate, t.InputSchema, t.UpdatedAt, t.ID)
	if err != nil {
		return apperrors.NewAppErrorf(apperrors.Internal, "updating http tool: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return apperrors.NewAppErrorf(apperrors.Internal, "checking update result: %v", err)
	}
	if rows == 0 {
		return apperrors.NewAppError(apperrors.NotFound, "http tool not found")
	}
	return nil
}

func (r *HttpToolRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM http_tools WHERE id = $1`, id)
	if err != nil {
		return apperrors.NewAppErrorf(apperrors.Internal, "deleting http tool: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return apperrors.NewAppErrorf(apperrors.Internal, "checking delete result: %v", err)
	}
	if rows == 0 {
		return apperrors.NewAppError(apperrors.NotFound, "http tool not found")
	}
	return nil
}

// scanner interface for both *sql.Rows and *sql.Row
type scanner interface {
	Scan(dest ...any) error
}

func scanFromRow(s scanner) (*httptool.HttpTool, error) {
	var t httptool.HttpTool
	var headersJSON string
	err := s.Scan(&t.ID, &t.Name, &t.Description, &t.Method, &t.URL, &headersJSON, &t.BodyTemplate, &t.InputSchema, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(headersJSON), &t.Headers); err != nil {
		return nil, apperrors.NewAppErrorf(apperrors.Internal, "unmarshaling headers: %v", err)
	}
	if t.Headers == nil {
		t.Headers = []httptool.HttpToolHeader{}
	}

	return &t, nil
}

func scanHttpTool(rows *sql.Rows) (*httptool.HttpTool, error) {
	t, err := scanFromRow(rows)
	if err != nil {
		return nil, apperrors.NewAppErrorf(apperrors.Internal, "scanning http tool: %v", err)
	}
	return t, nil
}

func scanHttpToolRow(row *sql.Row) (*httptool.HttpTool, error) {
	return scanFromRow(row)
}
