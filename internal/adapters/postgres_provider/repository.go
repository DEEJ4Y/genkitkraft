package postgresperson

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	apperrors "github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/provider"
	providerrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/provider_repo"
)

// Compile-time check that ProviderRepository implements the port interface.
var _ providerrepo.ProviderRepository = (*ProviderRepository)(nil)

// ProviderRepository implements providerrepo.ProviderRepository using PostgreSQL.
type ProviderRepository struct {
	db *sql.DB
}

// NewProviderRepository creates a new PostgreSQL-backed provider repository.
func NewProviderRepository(db *sql.DB) *ProviderRepository {
	return &ProviderRepository{db: db}
}

func (r *ProviderRepository) List(ctx context.Context) ([]*provider.Provider, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, provider_type, api_key, base_url, config, enabled, created_at, updated_at
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
		`SELECT id, name, provider_type, api_key, base_url, config, enabled, created_at, updated_at
		 FROM providers WHERE id = $1`, id)

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
		`SELECT id, name, provider_type, api_key, base_url, config, enabled, created_at, updated_at
		 FROM providers WHERE provider_type = $1`, string(pt))

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

	rawConfig := p.RawConfig
	if len(rawConfig) == 0 {
		rawConfig = json.RawMessage("{}")
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO providers (id, name, provider_type, api_key, base_url, config, enabled, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		p.ID, p.Name, string(p.ProviderType), nullString(p.APIKey), p.BaseURL,
		string(rawConfig), p.Enabled, p.CreatedAt, p.UpdatedAt)
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

	rawConfig := p.RawConfig
	if len(rawConfig) == 0 {
		rawConfig = json.RawMessage("{}")
	}

	result, err := r.db.ExecContext(ctx,
		`UPDATE providers SET name = $1, api_key = $2, base_url = $3, config = $4, enabled = $5, updated_at = $6
		 WHERE id = $7`,
		p.Name, nullString(p.APIKey), p.BaseURL, string(rawConfig),
		p.Enabled, p.UpdatedAt, p.ID)
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
	result, err := r.db.ExecContext(ctx, `DELETE FROM providers WHERE id = $1`, id)
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

func scanProvider(rows *sql.Rows) (*provider.Provider, error) {
	var p provider.Provider
	var apiKey sql.NullString
	var configStr string
	if err := rows.Scan(&p.ID, &p.Name, &p.ProviderType, &apiKey, &p.BaseURL, &configStr, &p.Enabled, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil, apperrors.NewAppErrorf(apperrors.Internal, "scanning provider: %v", err)
	}
	if apiKey.Valid {
		p.APIKey = &apiKey.String
	}
	p.RawConfig = json.RawMessage(configStr)

	cfg, err := provider.ParseProviderConfig(string(p.ProviderType), p.RawConfig)
	if err != nil {
		return nil, apperrors.NewAppErrorf(apperrors.Internal, "parsing provider config: %v", err)
	}
	p.Config = cfg

	return &p, nil
}

func scanProviderRow(row *sql.Row) (*provider.Provider, error) {
	var p provider.Provider
	var apiKey sql.NullString
	var configStr string
	if err := row.Scan(&p.ID, &p.Name, &p.ProviderType, &apiKey, &p.BaseURL, &configStr, &p.Enabled, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil, err // caller handles sql.ErrNoRows
	}
	if apiKey.Valid {
		p.APIKey = &apiKey.String
	}
	p.RawConfig = json.RawMessage(configStr)

	cfg, err := provider.ParseProviderConfig(string(p.ProviderType), p.RawConfig)
	if err != nil {
		return nil, apperrors.NewAppErrorf(apperrors.Internal, "parsing provider config: %v", err)
	}
	p.Config = cfg

	return &p, nil
}

func isUniqueConstraintError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func nullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}
