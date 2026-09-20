package mock

import (
	"context"

	"github.com/DEEJ4Y/genkitkraft/internal/domain/prompt"
	promptrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/prompt_repo"
)

// Compile-time check.
var _ promptrepo.PromptRepository = (*PromptRepository)(nil)

// PromptRepository is a mock implementation of the PromptRepository port for testing.
type PromptRepository struct {
	ListResult []*prompt.Prompt
	ListErr    error

	CountResult int
	CountErr    error

	GetByIDResult *prompt.Prompt
	GetByIDErr    error

	CreateErr error
	UpdateErr error
	DeleteErr error

	LastCreate *prompt.Prompt
	LastUpdate *prompt.Prompt
}

func (m *PromptRepository) List(_ context.Context, _, _ int) ([]*prompt.Prompt, error) {
	return m.ListResult, m.ListErr
}

func (m *PromptRepository) Count(_ context.Context) (int, error) {
	return m.CountResult, m.CountErr
}

func (m *PromptRepository) GetByID(_ context.Context, _ string) (*prompt.Prompt, error) {
	return m.GetByIDResult, m.GetByIDErr
}

func (m *PromptRepository) Create(_ context.Context, p *prompt.Prompt) error {
	if p.ID == "" {
		p.ID = "mock-prompt-id"
	}
	m.LastCreate = p
	return m.CreateErr
}

func (m *PromptRepository) Update(_ context.Context, p *prompt.Prompt) error {
	m.LastUpdate = p
	return m.UpdateErr
}

func (m *PromptRepository) Delete(_ context.Context, _ string) error {
	return m.DeleteErr
}
