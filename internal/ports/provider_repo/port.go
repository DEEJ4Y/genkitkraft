package providerrepo

import (
	"context"

	"github.com/DEEJ4Y/genkitkraft/internal/domain/provider"
)

// ProviderRepository defines the contract for provider credential persistence.
type ProviderRepository interface {
	List(ctx context.Context) ([]*provider.Provider, error)
	GetByID(ctx context.Context, id string) (*provider.Provider, error)
	GetByType(ctx context.Context, pt provider.ProviderType) (*provider.Provider, error)
	Create(ctx context.Context, p *provider.Provider) error
	Update(ctx context.Context, p *provider.Provider) error
	Delete(ctx context.Context, id string) error
}
