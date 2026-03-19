package decorators

import (
	"context"
	"time"

	"github.com/DEEJ4Y/genkitkraft/internal/app/executors"
	"github.com/rs/zerolog"
)

// loggingDecorator wraps an ExecutorWithReturn with request logging.
type loggingDecorator[Params, Result any] struct {
	base         executors.ExecutorWithReturn[Params, Result]
	executorName string
	logger       zerolog.Logger
}

func (d loggingDecorator[P, R]) Execute(ctx context.Context, params P) (R, error) {
	d.logger.Info().Str("executor", d.executorName).Msg("executing")

	start := time.Now()
	result, err := d.base.Execute(ctx, params)
	duration := time.Since(start)

	if err != nil {
		d.logger.Error().
			Str("executor", d.executorName).
			Dur("duration", duration).
			Err(err).
			Msg("failed")
	} else {
		d.logger.Info().
			Str("executor", d.executorName).
			Dur("duration", duration).
			Msg("completed")
	}
	return result, err
}

// ApplyLogging wraps an ExecutorWithReturn with request logging.
func ApplyLogging[Params, Result any](
	base executors.ExecutorWithReturn[Params, Result],
	executorName string,
	logger zerolog.Logger,
) executors.ExecutorWithReturn[Params, Result] {
	return loggingDecorator[Params, Result]{
		base:         base,
		executorName: executorName,
		logger:       logger,
	}
}

// loggingExecutorDecorator wraps an Executor (no return value) with request logging.
type loggingExecutorDecorator[Params any] struct {
	base         executors.Executor[Params]
	executorName string
	logger       zerolog.Logger
}

func (d loggingExecutorDecorator[P]) Execute(ctx context.Context, params P) error {
	d.logger.Info().Str("executor", d.executorName).Msg("executing")

	start := time.Now()
	err := d.base.Execute(ctx, params)
	duration := time.Since(start)

	if err != nil {
		d.logger.Error().
			Str("executor", d.executorName).
			Dur("duration", duration).
			Err(err).
			Msg("failed")
	} else {
		d.logger.Info().
			Str("executor", d.executorName).
			Dur("duration", duration).
			Msg("completed")
	}
	return err
}

// ApplyLoggingExecutor wraps an Executor (no return value) with request logging.
func ApplyLoggingExecutor[Params any](
	base executors.Executor[Params],
	executorName string,
	logger zerolog.Logger,
) executors.Executor[Params] {
	return loggingExecutorDecorator[Params]{
		base:         base,
		executorName: executorName,
		logger:       logger,
	}
}
