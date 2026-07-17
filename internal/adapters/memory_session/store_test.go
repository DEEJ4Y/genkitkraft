package memorysession_test

import (
	"context"
	"testing"
	"time"

	inmemorycache "github.com/DEEJ4Y/genkitkraft/internal/adapters/in_memory_cache"
	memorysession "github.com/DEEJ4Y/genkitkraft/internal/adapters/memory_session"
	"github.com/DEEJ4Y/genkitkraft/internal/common/errors"
)

func newStore(t *testing.T) *memorysession.MemoryStore {
	t.Helper()
	return memorysession.NewMemoryStore(inmemorycache.NewCache(time.Minute).Scope("session"))
}

func TestCreateThenValidate(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()

	token, err := s.Create(ctx, "alice")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if token == "" {
		t.Fatal("Create returned an empty token")
	}

	username, err := s.Validate(ctx, token)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if username != "alice" {
		t.Errorf("Validate = %q, want %q", username, "alice")
	}
}

func TestValidateUnknownTokenIsUnauthorized(t *testing.T) {
	s := newStore(t)

	_, err := s.Validate(context.Background(), "never-issued")
	assertUnauthorized(t, err)
}

func TestValidateAfterDeleteIsUnauthorized(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()

	token, err := s.Create(ctx, "alice")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := s.Delete(ctx, token); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err = s.Validate(ctx, token)
	assertUnauthorized(t, err)
}

func TestDeleteUnknownTokenIsNoOp(t *testing.T) {
	s := newStore(t)
	if err := s.Delete(context.Background(), "never-issued"); err != nil {
		t.Errorf("Delete on an unknown token: %v", err)
	}
}

func TestTokensAreDistinct(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()

	first, err := s.Create(ctx, "alice")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	second, err := s.Create(ctx, "alice")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if first == second {
		t.Error("two sessions for the same user share a token")
	}
}

// Sessions for different users must not be confusable, and deleting one must not
// affect the other.
func TestSessionsAreIndependent(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()

	alice, err := s.Create(ctx, "alice")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	bob, err := s.Create(ctx, "bob")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := s.Delete(ctx, alice); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	username, err := s.Validate(ctx, bob)
	if err != nil {
		t.Fatalf("Validate bob after deleting alice: %v", err)
	}
	if username != "bob" {
		t.Errorf("Validate = %q, want %q", username, "bob")
	}
}

func assertUnauthorized(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("want an error, got nil")
	}
	appErr, ok := errors.IsAppError(err)
	if !ok {
		t.Fatalf("want *errors.AppError, got %T: %v", err, err)
	}
	if appErr.Code() != errors.Unauthorized {
		t.Errorf("error code = %v, want %v", appErr.Code(), errors.Unauthorized)
	}
}
