package mysqlmcpserver

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	apperrors "github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	mcpserver "github.com/DEEJ4Y/genkitkraft/internal/domain/mcp_server"
	mcpserverrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/mcp_server_repo"
)

// Compile-time check that McpServerRepository implements the port interface.
var _ mcpserverrepo.McpServerRepository = (*McpServerRepository)(nil)

// McpServerRepository implements mcpserverrepo.McpServerRepository using MySQL.
type McpServerRepository struct {
	db *sql.DB
}

// NewMcpServerRepository creates a new MySQL-backed MCP server repository.
func NewMcpServerRepository(db *sql.DB) *McpServerRepository {
	return &McpServerRepository{db: db}
}

func (r *McpServerRepository) List(ctx context.Context, limit, offset int) ([]*mcpserver.McpServer, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, transport, url, headers, created_at, updated_at
		 FROM mcp_servers ORDER BY created_at DESC LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, apperrors.NewAppErrorf(apperrors.Internal, "listing mcp servers: %v", err)
	}
	defer rows.Close()

	var servers []*mcpserver.McpServer
	for rows.Next() {
		s, err := scanMcpServer(rows)
		if err != nil {
			return nil, err
		}
		servers = append(servers, s)
	}
	return servers, rows.Err()
}

func (r *McpServerRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM mcp_servers`).Scan(&count)
	if err != nil {
		return 0, apperrors.NewAppErrorf(apperrors.Internal, "counting mcp servers: %v", err)
	}
	return count, nil
}

func (r *McpServerRepository) GetByID(ctx context.Context, id string) (*mcpserver.McpServer, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, transport, url, headers, created_at, updated_at
		 FROM mcp_servers WHERE id = ?`, id)

	s, err := scanMcpServerRow(row)
	if err == sql.ErrNoRows {
		return nil, apperrors.NewAppError(apperrors.NotFound, "mcp server not found")
	}
	if err != nil {
		return nil, apperrors.NewAppErrorf(apperrors.Internal, "getting mcp server: %v", err)
	}
	return s, nil
}

func (r *McpServerRepository) Create(ctx context.Context, s *mcpserver.McpServer) error {
	s.ID = uuid.New().String()
	now := time.Now().UTC().Truncate(time.Microsecond)
	s.CreatedAt = now
	s.UpdatedAt = now

	headersJSON, err := json.Marshal(s.Headers)
	if err != nil {
		return apperrors.NewAppErrorf(apperrors.Internal, "marshaling headers: %v", err)
	}

	_, err = r.db.ExecContext(ctx,
		`INSERT INTO mcp_servers (id, name, transport, url, headers, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		s.ID, s.Name, s.Transport, s.URL, string(headersJSON), s.CreatedAt, s.UpdatedAt)
	if err != nil {
		return apperrors.NewAppErrorf(apperrors.Internal, "creating mcp server: %v", err)
	}
	return nil
}

func (r *McpServerRepository) Update(ctx context.Context, s *mcpserver.McpServer) error {
	s.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)

	headersJSON, err := json.Marshal(s.Headers)
	if err != nil {
		return apperrors.NewAppErrorf(apperrors.Internal, "marshaling headers: %v", err)
	}

	result, err := r.db.ExecContext(ctx,
		`UPDATE mcp_servers SET name = ?, transport = ?, url = ?, headers = ?, updated_at = ? WHERE id = ?`,
		s.Name, s.Transport, s.URL, string(headersJSON), s.UpdatedAt, s.ID)
	if err != nil {
		return apperrors.NewAppErrorf(apperrors.Internal, "updating mcp server: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return apperrors.NewAppErrorf(apperrors.Internal, "checking update result: %v", err)
	}
	if rows == 0 {
		return apperrors.NewAppError(apperrors.NotFound, "mcp server not found")
	}
	return nil
}

func (r *McpServerRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM mcp_servers WHERE id = ?`, id)
	if err != nil {
		return apperrors.NewAppErrorf(apperrors.Internal, "deleting mcp server: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return apperrors.NewAppErrorf(apperrors.Internal, "checking delete result: %v", err)
	}
	if rows == 0 {
		return apperrors.NewAppError(apperrors.NotFound, "mcp server not found")
	}
	return nil
}

// scanner interface for both *sql.Rows and *sql.Row
type scanner interface {
	Scan(dest ...any) error
}

func scanFromRow(s scanner) (*mcpserver.McpServer, error) {
	var srv mcpserver.McpServer
	var headersJSON string
	err := s.Scan(&srv.ID, &srv.Name, &srv.Transport, &srv.URL, &headersJSON, &srv.CreatedAt, &srv.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(headersJSON), &srv.Headers); err != nil {
		return nil, apperrors.NewAppErrorf(apperrors.Internal, "unmarshaling headers: %v", err)
	}
	if srv.Headers == nil {
		srv.Headers = []mcpserver.McpServerHeader{}
	}

	return &srv, nil
}

func scanMcpServer(rows *sql.Rows) (*mcpserver.McpServer, error) {
	s, err := scanFromRow(rows)
	if err != nil {
		return nil, apperrors.NewAppErrorf(apperrors.Internal, "scanning mcp server: %v", err)
	}
	return s, nil
}

func scanMcpServerRow(row *sql.Row) (*mcpserver.McpServer, error) {
	return scanFromRow(row)
}
