# Error Handling Strategy

Define a custom error type with error codes that map cleanly to transport-level status codes:

```go
// internal/common/errors/errors.go

type ErrorCode int

const (
    Internal     ErrorCode = iota
    InvalidInput
    NotFound
    Conflict
    Unauthorized
    Forbidden
    Timeout
    Unavailable
)

type AppError struct {
    original error
    code     ErrorCode
    message  string
}

func NewNotFound(msg string) error          { return &AppError{code: NotFound, message: msg} }
func NewInvalidInput(msg string) error      { return &AppError{code: InvalidInput, message: msg} }
func NewForbidden(msg string) error         { return &AppError{code: Forbidden, message: msg} }
func WrapInternal(err error, msg string) error {
    return &AppError{original: err, code: Internal, message: msg}
}
```

## Error Flow

1. **Domain/Application layer** returns `AppError` with appropriate codes.
2. **Error handler decorator** wraps unexpected (non-`AppError`) errors as `Internal`.
3. **Handler layer** maps `AppError` codes to transport status codes (gRPC codes, HTTP status codes).

This ensures that internal details never leak to clients while still preserving meaningful error context for logging and debugging.
