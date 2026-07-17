package errors

import "fmt"

type ErrorCode int

const (
	Internal ErrorCode = iota
	InvalidInput
	NotFound
	Conflict
	Unauthorized
	Forbidden
	TooManyRequests
	// Unavailable marks a dependency outage: the request could not be served, but
	// the client did nothing wrong and retrying the same request may well work.
	// Appended last on purpose — these are iota values, and Internal must stay 0 so
	// a zero-valued ErrorCode keeps meaning "internal error".
	Unavailable
)

type AppError struct {
	code    ErrorCode
	message string
}

func (e *AppError) Error() string {
	return e.message
}

func (e *AppError) Code() ErrorCode {
	return e.code
}

func NewAppError(code ErrorCode, message string) *AppError {
	return &AppError{code: code, message: message}
}

func NewAppErrorf(code ErrorCode, format string, args ...any) *AppError {
	return &AppError{code: code, message: fmt.Sprintf(format, args...)}
}

func IsAppError(err error) (*AppError, bool) {
	if appErr, ok := err.(*AppError); ok {
		return appErr, true
	}
	return nil, false
}

func HTTPStatusCode(code ErrorCode) int {
	switch code {
	case Internal:
		return 500
	case InvalidInput:
		return 400
	case Unauthorized:
		return 401
	case Forbidden:
		return 403
	case NotFound:
		return 404
	case Conflict:
		return 409
	case TooManyRequests:
		return 429
	case Unavailable:
		return 503
	default:
		// An unmapped code is a programming error, not an Internal one; 500 is the
		// safe answer either way.
		return 500
	}
}
