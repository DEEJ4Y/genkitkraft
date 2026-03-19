package sqliteprovider

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/google/uuid"

	apperrors "github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/provider"
	providerrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/provider_repo"
)

// Compile-time check that ProviderRepository implements the port interface.
var _ providerrepo.ProviderRepository = (*ProviderRepository)(nil)

// ProviderRepository implements providerrepo.ProviderRepository using SQLite.
type ProviderRepository struct {
	db *sql.DB
}

// NewProviderRepository creates a new SQLite-backed provider repository.
func NewProviderRepository(db *sql.DB) *ProviderRepository {
	return &ProviderRepository{db: db}
}

func (r *ProviderRepository) List(ctx context.Context) ([]*provider.Provider, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, provider_type, api_key, base_url, enabled, created_at, updated_at
		 FROM providers ORDER BY created_at`)
	if err != nil {
		return nil, apperrors.NewAppErrorf(apperrors.Internal, "listing providers: %v", err)
	}
	defer rows.Close()

	var providers []*provider.Provider
	for rows.Next() {
		p, err := scanProvider(rows)
		if err != nil {
			return nil, err
		}
		providers = append(providers, p)
	}
	return providers, rows.Err()
}

func (r *ProviderRepository) GetByID(ctx context.Context, id string) (*provider.Provider, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, provider_type, api_key, base_url, enabled, created_at, updated_at
		 FROM providers WHERE id = ?`, id)

	p, err := scanProviderRow(row)
	if err == sql.ErrNoRows {
		return nil, apperrors.NewAppError(apperrors.NotFound, "provider not found")
	}
	if err != nil {
		return nil, apperrors.NewAppErrorf(apperrors.Internal, "getting provider: %v", err)
	}
	return p, nil
}

func (r *ProviderRepository) GetByType(ctx context.Context, pt provider.ProviderType) (*provider.Provider, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, provider_type, api_key, base_url, enabled, created_at, updated_at
		 FROM providers WHERE provider_type = ?`, string(pt))

	p, err := scanProviderRow(row)
	if err == sql.ErrNoRows {
		return nil, apperrors.NewAppError(apperrors.NotFound, "provider not found")
	}
	if err != nil {
		return nil, apperrors.NewAppErrorf(apperrors.Internal, "getting provider by type: %v", err)
	}
	return p, nil
}

func (r *ProviderRepository) Create(ctx context.Context, p *provider.Provider) error {
	p.ID = uuid.New().String()
	now := time.Now().UTC()
	p.CreatedAt = now
	p.UpdatedAt = now

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO providers (id, name, provider_type, api_key, base_url, enabled, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.Name, string(p.ProviderType), p.APIKey, p.BaseURL, boolToInt(p.Enabled),
		p.CreatedAt, p.UpdatedAt)
	if err != nil {
		if isUniqueConstraintError(err) {
			return apperrors.NewAppErrorf(apperrors.Conflict, "a provider of type %q already exists", p.ProviderType)
		}
		return apperrors.NewAppErrorf(apperrors.Internal, "creating provider: %v", err)
	}
	return nil
}

func (r *ProviderRepository) Update(ctx context.Context, p *provider.Provider) error {
	p.UpdatedAt = time.Now().UTC()

	result, err := r.db.ExecContext(ctx,
		`UPDATE providers SET name = ?, api_key = ?, base_url = ?, enabled = ?, updated_at = ?
		 WHERE id = ?`,
		p.Name, p.APIKey, p.BaseURL, boolToInt(p.Enabled), p.UpdatedAt, p.ID)
	if err != nil {
		if isUniqueConstraintError(err) {
			return apperrors.NewAppError(apperrors.Conflict, "provider type conflict")
		}
		return apperrors.NewAppErrorf(apperrors.Internal, "updating provider: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return apperrors.NewAppErrorf(apperrors.Internal, "checking update result: %v", err)
	}
	if rows == 0 {
		return apperrors.NewAppError(apperrors.NotFound, "provider not found")
	}
	return nil
}

func (r *ProviderRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM providers WHERE id = ?`, id)
	if err != nil {
		return apperrors.NewAppErrorf(apperrors.Internal, "deleting provider: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return apperrors.NewAppErrorf(apperrors.Internal, "checking delete result: %v", err)
	}
	if rows == 0 {
		return apperrors.NewAppError(apperrors.NotFound, "provider not found")
	}
	return nil
}

// scanner is satisfied by both *sql.Rows and *sql.Row.
type scanner interface {
	Scan(dest ...any) error
}

func scanProvider(rows *sql.Rows) (*provider.Provider, error) {
	var p provider.Provider
	var enabled int
	if err := rows.Scan(&p.ID, &p.Name, &p.ProviderType, &p.APIKey, &p.BaseURL, &enabled, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil, apperrors.NewAppErrorf(apperrors.Internal, "scanning provider: %v", err)
	}
	p.Enabled = enabled != 0
	return &p, nil
}

func scanProviderRow(row *sql.Row) (*provider.Provider, error) {
	var p provider.Provider
	var enabled int
	if err := row.Scan(&p.ID, &p.Name, &p.ProviderType, &p.APIKey, &p.BaseURL, &enabled, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil, err // caller handles sql.ErrNoRows
	}
	p.Enabled = enabled != 0
	return &p, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func isUniqueConstraintError(err error) bool {
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}
