# Cross-Cutting Concerns with the Decorator Pattern

Instead of scattering logging, tracing, and caching logic throughout use cases, wrap them with **decorators**. Decorators implement the same `Executor` interface and add behavior around the real implementation.

## Logging Decorator

```go
// internal/app/decorators/logging.go

type loggingDecorator[Params, Result any] struct {
    base         ExecutorWithReturn[Params, Result]
    executorName string
    logger       zerolog.Logger
}

func (d loggingDecorator[P, R]) Execute(ctx context.Context, params P) (R, error) {
    d.logger.Info().Str("executor", d.executorName).Interface("input", params).Msg("executing")

    result, err := d.base.Execute(ctx, params)

    if err != nil {
        d.logger.Error().Str("executor", d.executorName).Err(err).Msg("failed")
    } else {
        d.logger.Info().Str("executor", d.executorName).Msg("completed")
    }
    return result, err
}
```

## Tracing Decorator

```go
// internal/app/decorators/tracing.go

type tracingDecorator[Params, Result any] struct {
    base         ExecutorWithReturn[Params, Result]
    executorName string
    tracer       trace.Tracer
}

func (d tracingDecorator[P, R]) Execute(ctx context.Context, params P) (R, error) {
    ctx, span := d.tracer.Start(ctx, fmt.Sprintf("executor:%s", d.executorName))
    defer span.End()

    result, err := d.base.Execute(ctx, params)
    if err != nil {
        span.RecordError(err)
    }
    return result, err
}
```

## Cache Invalidation Decorator

```go
// internal/app/decorators/cache_invalidation.go

type cacheInvalidationDecorator[Params, Result any] struct {
    base         ExecutorWithReturn[Params, Result]
    invalidators func(Params, Result) []CacheInvalidator
}

func (d cacheInvalidationDecorator[P, R]) Execute(ctx context.Context, params P) (R, error) {
    result, err := d.base.Execute(ctx, params)
    if err != nil {
        return result, err
    }

    // Async invalidation — don't block the response
    go func() {
        for _, inv := range d.invalidators(params, result) {
            inv.Delete(context.Background())
        }
    }()

    return result, nil
}
```

## Strict Error Handler Decorator

```go
// internal/app/decorators/error_handler.go

type errorHandlerDecorator[Params, Result any] struct {
    base         ExecutorWithReturn[Params, Result]
    executorName string
}

func (d errorHandlerDecorator[P, R]) Execute(ctx context.Context, params P) (R, error) {
    result, err := d.base.Execute(ctx, params)
    if err != nil {
        var appErr *errors.AppError
        if !errors.As(err, &appErr) {
            // Wrap unexpected errors to prevent leaking internals
            err = errors.WrapInternal(err, "internal error")
        }
    }
    return result, err
}
```

## Applying Decorators

Decorators are composed during construction. **Order matters** — the outermost decorator executes first:

```go
func NewCreateOrderCommand(repo repository.OrderRepository, logger zerolog.Logger) ExecutorWithReturn[CreateOrder, domain.ID] {
    cmd := &CreateOrderCommand{repo: repo}

    // Innermost → outermost
    var decorated ExecutorWithReturn[CreateOrder, domain.ID] = cmd
    decorated = ApplyTracing(decorated, "CreateOrder")
    decorated = ApplyErrorHandler(decorated, "CreateOrder")
    decorated = ApplyLogging(decorated, "CreateOrder", logger)

    return decorated
}
```

Execution order: **Logging → Error Handler → Tracing → Actual Command**
