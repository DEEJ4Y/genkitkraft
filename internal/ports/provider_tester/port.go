package providertester

import (
	"context"

	"github.com/DEEJ4Y/genkitkraft/internal/domain/provider"
)

// Tester defines the contract for testing provider API connectivity.
type Tester interface {
	Test(ctx context.Context, p *provider.Provider) (success bool, message string, err error)
}
