package promptrepo

import (
	"context"

	"github.com/DEEJ4Y/genkitkraft/internal/domain/prompt"
)

// PromptRepository defines the contract for prompt persistence.
type PromptRepository interface {
	List(ctx context.Context, limit, offset int) ([]*prompt.Prompt, error)
	Count(ctx context.Context) (int, error)
	GetByID(ctx context.Context, id string) (*prompt.Prompt, error)
	Create(ctx context.Context, p *prompt.Prompt) error
	Update(ctx context.Context, p *prompt.Prompt) error
	Delete(ctx context.Context, id string) error
}
