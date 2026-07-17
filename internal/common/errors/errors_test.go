package errors_test

import (
	"testing"

	"github.com/DEEJ4Y/genkitkraft/internal/common/errors"
)

func TestHTTPStatusCode(t *testing.T) {
	tests := []struct {
		name string
		code errors.ErrorCode
		want int
	}{
		{"Internal", errors.Internal, 500},
		{"InvalidInput", errors.InvalidInput, 400},
		{"Unauthorized", errors.Unauthorized, 401},
		{"Forbidden", errors.Forbidden, 403},
		{"NotFound", errors.NotFound, 404},
		{"Conflict", errors.Conflict, 409},
		{"TooManyRequests", errors.TooManyRequests, 429},
		{"Unavailable", errors.Unavailable, 503},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := errors.HTTPStatusCode(tt.code); got != tt.want {
				t.Errorf("HTTPStatusCode(%s) = %d, want %d", tt.name, got, tt.want)
			}
		})
	}
}

// Internal is the iota zero value, so a zero-valued ErrorCode — an AppError built
// without one, say — must keep meaning "internal error". Reordering the const
// block would silently remap every code in the codebase.
func TestInternalIsTheZeroValue(t *testing.T) {
	var zero errors.ErrorCode
	if zero != errors.Internal {
		t.Errorf("zero ErrorCode = %v, want Internal", zero)
	}
}

// An unmapped code must not fall through to a success status.
func TestUnmappedCodeIsServerError(t *testing.T) {
	if got := errors.HTTPStatusCode(errors.ErrorCode(999)); got != 500 {
		t.Errorf("HTTPStatusCode(unmapped) = %d, want 500", got)
	}
}
