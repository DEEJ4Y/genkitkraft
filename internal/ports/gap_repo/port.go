package gaprepo

import (
	"context"

	"github.com/DEEJ4Y/genkitkraft/internal/domain/gap"
)

// GapRepository defines the contract for gap persistence.
type GapRepository interface {
	List(ctx context.Context, agentID string, limit, offset int) ([]*gap.Gap, error)
	Count(ctx context.Context, agentID string) (int, error)
	GetByID(ctx context.Context, id string) (*gap.Gap, error)
	Create(ctx context.Context, g *gap.Gap) error
	Update(ctx context.Context, g *gap.Gap) error

	AddReference(ctx context.Context, ref *gap.Reference) error
	ListReferences(ctx context.Context, gapID string) ([]*gap.Reference, error)
}
