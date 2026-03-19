package executors

import "context"

// Executor defines a use case that performs an action without returning a value.
type Executor[Params any] interface {
	Execute(ctx context.Context, params Params) error
}

// ExecutorWithReturn defines a use case that performs an action and returns a result.
type ExecutorWithReturn[Params, Result any] interface {
	Execute(ctx context.Context, params Params) (Result, error)
}
