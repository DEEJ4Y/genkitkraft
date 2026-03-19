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
	default:
		return 500
	}
}
