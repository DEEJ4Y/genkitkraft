package mock

import (
	"context"

	"github.com/DEEJ4Y/genkitkraft/internal/domain/provider"
	providerrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/provider_repo"
)

// Compile-time check.
var _ providerrepo.ProviderRepository = (*ProviderRepository)(nil)

// ProviderRepository is a mock implementation of the ProviderRepository port for testing.
type ProviderRepository struct {
	ListResult []*provider.Provider
	ListErr    error

	GetByIDResult   *provider.Provider
	GetByIDErr      error
	GetByTypeResult *provider.Provider
	GetByTypeErr    error

	CreateErr error
	UpdateErr error
	DeleteErr error

	LastCreate *provider.Provider
	LastUpdate *provider.Provider
}

func (m *ProviderRepository) List(_ context.Context) ([]*provider.Provider, error) {
	return m.ListResult, m.ListErr
}

func (m *ProviderRepository) GetByID(_ context.Context, _ string) (*provider.Provider, error) {
	return m.GetByIDResult, m.GetByIDErr
}

func (m *ProviderRepository) GetByType(_ context.Context, _ provider.ProviderType) (*provider.Provider, error) {
	return m.GetByTypeResult, m.GetByTypeErr
}

func (m *ProviderRepository) Create(_ context.Context, p *provider.Provider) error {
	if p.ID == "" {
		p.ID = "mock-provider-id"
	}
	m.LastCreate = p
	return m.CreateErr
}

func (m *ProviderRepository) Update(_ context.Context, p *provider.Provider) error {
	m.LastUpdate = p
	return m.UpdateErr
}

func (m *ProviderRepository) Delete(_ context.Context, _ string) error {
	return m.DeleteErr
}
