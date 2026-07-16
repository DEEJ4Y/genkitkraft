package decorators_test

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/rs/zerolog"

	inmemorycache "github.com/DEEJ4Y/genkitkraft/internal/adapters/in_memory_cache"
	"github.com/DEEJ4Y/genkitkraft/internal/app/commands"
	"github.com/DEEJ4Y/genkitkraft/internal/app/decorators"
	"github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	"github.com/DEEJ4Y/genkitkraft/internal/ports/cache"
)

const testIP = "203.0.113.7"

// stubLogin stands in for the real login command: it returns whatever the test
// tells it to, and counts how often it was actually reached so tests can assert
// that a rate-limited request never hit the inner command.
type stubLogin struct {
	err   error
	panic bool
	calls int
}

func (s *stubLogin) Execute(_ context.Context, params commands.LoginParams) (commands.LoginResult, error) {
	s.calls++
	if s.panic {
		panic("boom")
	}
	if s.err != nil {
		return commands.LoginResult{}, s.err
	}
	return commands.LoginResult{Token: "token", Username: params.Username}, nil
}

func newDecorator(inner *stubLogin, c cache.Cache) *decorators.RateLimitingLoginDecorator {
	return decorators.NewRateLimitingLoginDecorator(inner, c, zerolog.New(io.Discard))
}

func newSharedCache(t *testing.T) cache.Store {
	t.Helper()
	return inmemorycache.NewCache(time.Minute)
}

func badCredentials() error {
	return errors.NewAppError(errors.Unauthorized, "invalid credentials")
}

func login(t *testing.T, d *decorators.RateLimitingLoginDecorator) error {
	t.Helper()
	_, err := d.Execute(context.Background(), commands.LoginParams{
		Username: "alice",
		Password: "wrong",
		ClientIP: testIP,
	})
	return err
}

func assertTooManyRequests(t *testing.T, err error) {
	t.Helper()
	appErr, ok := errors.IsAppError(err)
	if !ok {
		t.Fatalf("want *errors.AppError, got %T: %v", err, err)
	}
	if appErr.Code() != errors.TooManyRequests {
		t.Errorf("error code = %v, want %v (TooManyRequests)", appErr.Code(), errors.TooManyRequests)
	}
}

func TestAllowsUpToTheLimitThenBlocks(t *testing.T) {
	inner := &stubLogin{err: badCredentials()}
	d := newDecorator(inner, newSharedCache(t).Scope("rate_limit"))

	// The first 5 failures must reach the login command.
	for i := 1; i <= 5; i++ {
		if err := login(t, d); err == nil {
			t.Fatalf("attempt %d: want a credentials error, got nil", i)
		}
	}
	if inner.calls != 5 {
		t.Fatalf("inner called %d times over 5 attempts, want 5", inner.calls)
	}

	// The 6th is rejected before reaching it.
	assertTooManyRequests(t, login(t, d))
	if inner.calls != 5 {
		t.Errorf("inner called %d times after a rate-limited attempt, want it untouched at 5", inner.calls)
	}
}

// This is the multi-instance bypass from issue #29: two decorators sharing one
// backing cache stand in for two instances behind a load balancer. Spreading
// attempts across them must not buy the attacker extra tries.
func TestLimitIsSharedAcrossInstances(t *testing.T) {
	store := newSharedCache(t)
	innerA, innerB := &stubLogin{err: badCredentials()}, &stubLogin{err: badCredentials()}
	instanceA := newDecorator(innerA, store.Scope("rate_limit"))
	instanceB := newDecorator(innerB, store.Scope("rate_limit"))

	// 4 failures on A, 1 on B — 5 total, all still allowed through.
	for i := 1; i <= 4; i++ {
		if err := login(t, instanceA); err == nil {
			t.Fatalf("instance A attempt %d: want a credentials error, got nil", i)
		}
	}
	if err := login(t, instanceB); err == nil {
		t.Fatal("instance B attempt 5: want a credentials error, got nil")
	}

	// The 6th must be blocked on either instance, whichever the balancer picks.
	assertTooManyRequests(t, login(t, instanceB))
	assertTooManyRequests(t, login(t, instanceA))

	if innerA.calls+innerB.calls != 5 {
		t.Errorf("inner reached %d times across both instances, want 5", innerA.calls+innerB.calls)
	}
}

func TestSuccessfulLoginClearsTheCount(t *testing.T) {
	inner := &stubLogin{err: badCredentials()}
	d := newDecorator(inner, newSharedCache(t).Scope("rate_limit"))

	for i := 1; i <= 4; i++ {
		if err := login(t, d); err == nil {
			t.Fatalf("attempt %d: want a credentials error, got nil", i)
		}
	}

	inner.err = nil
	if _, err := d.Execute(context.Background(), commands.LoginParams{Username: "alice", ClientIP: testIP}); err != nil {
		t.Fatalf("successful login: %v", err)
	}

	// The count is reset, so a fresh run of failures is allowed again.
	inner.err = badCredentials()
	for i := 1; i <= 5; i++ {
		if err := login(t, d); err == nil {
			t.Fatalf("post-success attempt %d: want a credentials error, got nil", i)
		} else {
			assertNotRateLimited(t, err, i)
		}
	}
}

func assertNotRateLimited(t *testing.T, err error, attempt int) {
	t.Helper()
	if appErr, ok := errors.IsAppError(err); ok && appErr.Code() == errors.TooManyRequests {
		t.Fatalf("attempt %d was rate limited, but a successful login should have reset the count", attempt)
	}
}

func TestLimitIsPerIP(t *testing.T) {
	inner := &stubLogin{err: badCredentials()}
	d := newDecorator(inner, newSharedCache(t).Scope("rate_limit"))
	ctx := context.Background()

	for i := 1; i <= 5; i++ {
		if _, err := d.Execute(ctx, commands.LoginParams{Username: "alice", ClientIP: "198.51.100.1"}); err == nil {
			t.Fatalf("attempt %d: want a credentials error, got nil", i)
		}
	}

	// A different IP is unaffected by the first one's exhausted budget.
	_, err := d.Execute(ctx, commands.LoginParams{Username: "alice", ClientIP: "198.51.100.2"})
	if appErr, ok := errors.IsAppError(err); ok && appErr.Code() == errors.TooManyRequests {
		t.Error("an unrelated IP was rate limited by another IP's failures")
	}
}

// A panic in the inner command must release only the attempt it consumed, so a
// transient bug cannot lock an IP out.
func TestPanicReleasesItsAttempt(t *testing.T) {
	inner := &stubLogin{err: badCredentials()}
	d := newDecorator(inner, newSharedCache(t).Scope("rate_limit"))

	for i := 1; i <= 4; i++ {
		if err := login(t, d); err == nil {
			t.Fatalf("attempt %d: want a credentials error, got nil", i)
		}
	}

	inner.panic = true
	func() {
		defer func() {
			if recover() == nil {
				t.Error("want the panic to propagate, got nil")
			}
		}()
		_, _ = d.Execute(context.Background(), commands.LoginParams{Username: "alice", ClientIP: testIP})
	}()
	inner.panic = false

	// The count is back to 4, so one more attempt is allowed and the next blocked.
	if err := login(t, d); err == nil {
		t.Fatal("attempt after the panic: want a credentials error, got nil")
	} else {
		assertNotRateLimited(t, err, 5)
	}
	assertTooManyRequests(t, login(t, d))
}

// failingCache stands in for a cache backend that is down.
type failingCache struct{ err error }

func (f *failingCache) Get(context.Context, string) (string, bool) { return "", false }
func (f *failingCache) Set(context.Context, string, string, time.Duration) error {
	return f.err
}
func (f *failingCache) Delete(context.Context, string) error { return f.err }
func (f *failingCache) Increment(context.Context, string, time.Duration) (int64, error) {
	return 0, f.err
}
func (f *failingCache) Decrement(context.Context, string) error { return f.err }

// With the cache down, sessions cannot be validated anyway, so the limiter fails
// closed rather than handing out unlimited attempts.
func TestCacheFailureDeniesLogin(t *testing.T) {
	inner := &stubLogin{}
	d := newDecorator(inner, &failingCache{err: io.ErrUnexpectedEOF})

	assertTooManyRequests(t, login(t, d))
	if inner.calls != 0 {
		t.Errorf("inner called %d times while the cache was down, want 0", inner.calls)
	}
}
