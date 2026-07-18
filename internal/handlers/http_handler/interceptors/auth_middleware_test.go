package interceptors_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DEEJ4Y/genkitkraft/internal/app"
	"github.com/DEEJ4Y/genkitkraft/internal/app/queries"
	"github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	"github.com/DEEJ4Y/genkitkraft/internal/handlers/http_handler/interceptors"
)

type stubGetMe struct{ err error }

func (s stubGetMe) Execute(context.Context, queries.GetMeParams) (queries.GetMeResult, error) {
	if s.err != nil {
		return queries.GetMeResult{}, s.err
	}
	return queries.GetMeResult{Username: "alice"}, nil
}

type stubAuthStatus struct{}

func (stubAuthStatus) Execute(context.Context, queries.GetAuthStatusParams) (queries.GetAuthStatusResult, error) {
	return queries.GetAuthStatusResult{Required: true}, nil
}

// serve runs a request with a session cookie through the middleware and reports
// the status the middleware produced. The wrapped handler writes 200, so any other
// status is the middleware's own answer.
func serve(t *testing.T, getMeErr error) int {
	t.Helper()

	authApp := &app.AuthApp{
		Queries: app.AuthQueries{
			GetMe:         stubGetMe{err: getMeErr},
			GetAuthStatus: stubAuthStatus{},
		},
	}
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: "some-token"})
	rec := httptest.NewRecorder()

	interceptors.AuthMiddleware(authApp)(next).ServeHTTP(rec, req)
	return rec.Code
}

// The point of the change: during a cache outage every authenticated route used to
// answer 401 while login answered 503, so the user was bounced between a "you are
// logged out" that sent them to a login page that then said "we are broken". This
// middleware — not the session store — is what emitted those 401s, so without it
// honouring the code the session store's Unavailable never reaches the client.
func TestAuthMiddlewareReportsCacheOutageAs503(t *testing.T) {
	got := serve(t, errors.NewAppError(errors.Unavailable, "session store is temporarily unavailable, try again later"))

	if got != http.StatusServiceUnavailable {
		t.Errorf("status during a cache outage = %d, want %d", got, http.StatusServiceUnavailable)
	}
}

// A genuine expired or unknown session is still a logout, and must stay a 401.
func TestAuthMiddlewareReportsInvalidSessionAs401(t *testing.T) {
	got := serve(t, errors.NewAppError(errors.Unauthorized, "invalid or expired session"))

	if got != http.StatusUnauthorized {
		t.Errorf("status for an invalid session = %d, want %d", got, http.StatusUnauthorized)
	}
}

// Everything that is not a recognised outage must fail closed at the auth gate.
// This is the guard against mapping the code straight through: a session lookup
// returning some other AppError must not leak its status to an unauthenticated
// caller, and a bare error must not surface as a 500 from the gate.
func TestAuthMiddlewareFailsClosedOnUnexpectedError(t *testing.T) {
	for _, tt := range []struct {
		name string
		err  error
	}{
		{"bare error", io.ErrUnexpectedEOF},
		{"unrelated app error", errors.NewAppError(errors.NotFound, "nope")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := serve(t, tt.err); got != http.StatusUnauthorized {
				t.Errorf("status = %d, want %d", got, http.StatusUnauthorized)
			}
		})
	}
}

// A request with no cookie never reaches the session store, so it stays a 401.
func TestAuthMiddlewareRejectsMissingCookieAs401(t *testing.T) {
	authApp := &app.AuthApp{
		Queries: app.AuthQueries{
			GetMe:         stubGetMe{},
			GetAuthStatus: stubAuthStatus{},
		},
	}
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rec := httptest.NewRecorder()
	interceptors.AuthMiddleware(authApp)(next).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/auth/me", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status without a session cookie = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}
