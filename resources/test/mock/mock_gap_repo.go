package mock

import (
	"context"

	"github.com/DEEJ4Y/genkitkraft/internal/domain/gap"
	gaprepo "github.com/DEEJ4Y/genkitkraft/internal/ports/gap_repo"
)

// Compile-time check.
var _ gaprepo.GapRepository = (*GapRepository)(nil)

// GapRepository is a mock implementation of the GapRepository port for testing.
type GapRepository struct {
	// ListGaps is returned by List(). Set before calling.
	ListGaps []*gap.Gap
	ListErr  error

	CountResult int
	CountErr    error

	GetByIDResult *gap.Gap
	GetByIDErr    error

	CreateErr error
	UpdateErr error

	AddReferenceErr error

	// References is returned by ListReferences(). Set before calling.
	References    []*gap.Reference
	ReferencesErr error

	// LastCreate/LastUpdate/LastAddReference capture the most recent call's
	// argument for assertions.
	LastCreate       *gap.Gap
	LastUpdate       *gap.Gap
	LastAddReference *gap.Reference

	LastListAgentID string
	LastListStatus  gap.Status
	LastListLimit   int
	LastListOffset  int

	LastCountAgentID        string
	LastCountStatus         gap.Status
	LastGetByID             string
	LastListReferencesGapID string
}

func (m *GapRepository) List(_ context.Context, agentID string, status gap.Status, limit, offset int) ([]*gap.Gap, error) {
	m.LastListAgentID = agentID
	m.LastListStatus = status
	m.LastListLimit = limit
	m.LastListOffset = offset
	return m.ListGaps, m.ListErr
}

func (m *GapRepository) Count(_ context.Context, agentID string, status gap.Status) (int, error) {
	m.LastCountAgentID = agentID
	m.LastCountStatus = status
	return m.CountResult, m.CountErr
}

func (m *GapRepository) GetByID(_ context.Context, id string) (*gap.Gap, error) {
	m.LastGetByID = id
	return m.GetByIDResult, m.GetByIDErr
}

func (m *GapRepository) Create(_ context.Context, g *gap.Gap) error {
	if g.ID == "" {
		g.ID = "mock-gap-id"
	}
	m.LastCreate = g
	return m.CreateErr
}

func (m *GapRepository) Update(_ context.Context, g *gap.Gap) error {
	m.LastUpdate = g
	return m.UpdateErr
}

func (m *GapRepository) AddReference(_ context.Context, ref *gap.Reference) error {
	m.LastAddReference = ref
	return m.AddReferenceErr
}

func (m *GapRepository) ListReferences(_ context.Context, gapID string) ([]*gap.Reference, error) {
	m.LastListReferencesGapID = gapID
	return m.References, m.ReferencesErr
}
