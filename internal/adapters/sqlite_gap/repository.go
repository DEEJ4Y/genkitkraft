package sqlitegap

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	apperrors "github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/gap"
	gaprepo "github.com/DEEJ4Y/genkitkraft/internal/ports/gap_repo"
)

// Compile-time check that GapRepository implements the port interface.
var _ gaprepo.GapRepository = (*GapRepository)(nil)

// GapRepository implements gaprepo.GapRepository using SQLite.
type GapRepository struct {
	db *sql.DB
}

// NewGapRepository creates a new SQLite-backed gap repository.
func NewGapRepository(db *sql.DB) *GapRepository {
	return &GapRepository{db: db}
}

func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func scanGap(row interface{ Scan(dest ...any) error }) (*gap.Gap, error) {
	var g gap.Gap
	var suggestedResolution, dismissalCategory, dismissalReason sql.NullString
	err := row.Scan(&g.ID, &g.AgentID, &g.Category, &g.Context, &g.Details, &suggestedResolution,
		&g.Status, &dismissalCategory, &dismissalReason, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		return nil, err
	}
	g.SuggestedResolution = suggestedResolution.String
	g.DismissalCategory = dismissalCategory.String
	g.DismissalReason = dismissalReason.String
	return &g, nil
}

const gapColumns = `id, agent_id, category, context, details, suggested_resolution, status, dismissal_category, dismissal_reason, created_at, updated_at`

func (r *GapRepository) List(ctx context.Context, agentID string, limit, offset int) ([]*gap.Gap, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+gapColumns+`
		 FROM agent_gaps WHERE agent_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`, agentID, limit, offset)
	if err != nil {
		return nil, apperrors.NewAppErrorf(apperrors.Internal, "listing gaps: %v", err)
	}
	defer rows.Close()

	var gaps []*gap.Gap
	for rows.Next() {
		g, err := scanGap(rows)
		if err != nil {
			return nil, apperrors.NewAppErrorf(apperrors.Internal, "scanning gap: %v", err)
		}
		gaps = append(gaps, g)
	}
	return gaps, rows.Err()
}

func (r *GapRepository) Count(ctx context.Context, agentID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM agent_gaps WHERE agent_id = ?`, agentID).Scan(&count)
	if err != nil {
		return 0, apperrors.NewAppErrorf(apperrors.Internal, "counting gaps: %v", err)
	}
	return count, nil
}

func (r *GapRepository) GetByID(ctx context.Context, id string) (*gap.Gap, error) {
	row := r.db.QueryRowContext(ctx, `SELECT `+gapColumns+` FROM agent_gaps WHERE id = ?`, id)
	g, err := scanGap(row)
	if err == sql.ErrNoRows {
		return nil, apperrors.NewAppError(apperrors.NotFound, "gap not found")
	}
	if err != nil {
		return nil, apperrors.NewAppErrorf(apperrors.Internal, "getting gap: %v", err)
	}
	return g, nil
}

func (r *GapRepository) Create(ctx context.Context, g *gap.Gap) error {
	g.ID = uuid.New().String()
	now := time.Now().UTC()
	g.CreatedAt = now
	g.UpdatedAt = now
	if g.Status == "" {
		g.Status = gap.StatusOpen
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO agent_gaps (id, agent_id, category, context, details, suggested_resolution, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		g.ID, g.AgentID, g.Category, g.Context, g.Details, nullableString(g.SuggestedResolution), g.Status, g.CreatedAt, g.UpdatedAt)
	if err != nil {
		return apperrors.NewAppErrorf(apperrors.Internal, "creating gap: %v", err)
	}
	return nil
}

func (r *GapRepository) Update(ctx context.Context, g *gap.Gap) error {
	g.UpdatedAt = time.Now().UTC()

	result, err := r.db.ExecContext(ctx,
		`UPDATE agent_gaps SET category = ?, context = ?, details = ?, suggested_resolution = ?,
		 status = ?, dismissal_category = ?, dismissal_reason = ?, updated_at = ? WHERE id = ?`,
		g.Category, g.Context, g.Details, nullableString(g.SuggestedResolution),
		g.Status, nullableString(g.DismissalCategory), nullableString(g.DismissalReason), g.UpdatedAt, g.ID)
	if err != nil {
		return apperrors.NewAppErrorf(apperrors.Internal, "updating gap: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return apperrors.NewAppErrorf(apperrors.Internal, "checking update result: %v", err)
	}
	if rows == 0 {
		return apperrors.NewAppError(apperrors.NotFound, "gap not found")
	}
	return nil
}

func (r *GapRepository) AddReference(ctx context.Context, ref *gap.Reference) error {
	ref.ID = uuid.New().String()
	ref.CreatedAt = time.Now().UTC()

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO agent_gap_references (id, gap_id, session_id, message_id, created_at) VALUES (?, ?, ?, ?, ?)`,
		ref.ID, ref.GapID, nullableString(ref.SessionID), nullableString(ref.MessageID), ref.CreatedAt)
	if err != nil {
		return apperrors.NewAppErrorf(apperrors.Internal, "adding gap reference: %v", err)
	}
	return nil
}

func (r *GapRepository) ListReferences(ctx context.Context, gapID string) ([]*gap.Reference, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, gap_id, session_id, message_id, created_at
		 FROM agent_gap_references WHERE gap_id = ? ORDER BY created_at ASC`, gapID)
	if err != nil {
		return nil, apperrors.NewAppErrorf(apperrors.Internal, "listing gap references: %v", err)
	}
	defer rows.Close()

	var refs []*gap.Reference
	for rows.Next() {
		var ref gap.Reference
		var sessionID, messageID sql.NullString
		if err := rows.Scan(&ref.ID, &ref.GapID, &sessionID, &messageID, &ref.CreatedAt); err != nil {
			return nil, apperrors.NewAppErrorf(apperrors.Internal, "scanning gap reference: %v", err)
		}
		ref.SessionID = sessionID.String
		ref.MessageID = messageID.String
		refs = append(refs, &ref)
	}
	return refs, rows.Err()
}
