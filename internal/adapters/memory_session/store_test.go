package memorysession_test

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/rs/zerolog"

	inmemorycache "github.com/DEEJ4Y/genkitkraft/internal/adapters/in_memory_cache"
	memorysession "github.com/DEEJ4Y/genkitkraft/internal/adapters/memory_session"
	"github.com/DEEJ4Y/genkitkraft/internal/common/errors"
)

func newStore(t *testing.T) *memorysession.MemoryStore {
	t.Helper()
	return memorysession.NewMemoryStore(inmemorycache.NewCache(time.Minute, zerolog.New(io.Discard)).Scope("session"))
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

// failingCache stands in for a cache backend that is down.
type failingCache struct{ err error }

func (f *failingCache) Get(context.Context, string) (string, bool, error) { return "", false, f.err }
func (f *failingCache) Set(context.Context, string, string, time.Duration) error {
	return f.err
}
func (f *failingCache) Delete(context.Context, string) error { return f.err }
func (f *failingCache) Increment(context.Context, string, time.Duration) (int64, error) {
	return 0, f.err
}
func (f *failingCache) Decrement(context.Context, string) error { return f.err }

// A cache outage must not read as an expired session. Unauthorized would tell the
// user to log in again — which cannot work while the store backing sessions is
// down, and login itself answers 503 — so the API would describe two different
// failures at once. The tests above pin the other side of this: a genuine miss is
// still a logout, so the error path must not swallow them.
func TestValidateReportsCacheOutageAsUnavailable(t *testing.T) {
	s := memorysession.NewMemoryStore(&failingCache{err: io.ErrUnexpectedEOF})

	_, err := s.Validate(context.Background(), "some-token")

	assertUnavailable(t, err)
}

func assertUnauthorized(t *testing.T, err error) {
	t.Helper()
	assertCode(t, err, errors.Unauthorized)
}

func assertUnavailable(t *testing.T, err error) {
	t.Helper()
	assertCode(t, err, errors.Unavailable)
}

func assertCode(t *testing.T, err error, want errors.ErrorCode) {
	t.Helper()
	if err == nil {
		t.Fatal("want an error, got nil")
	}
	appErr, ok := errors.IsAppError(err)
	if !ok {
		t.Fatalf("want *errors.AppError, got %T: %v", err, err)
	}
	if appErr.Code() != want {
		t.Errorf("error code = %v, want %v", appErr.Code(), want)
	}
}
