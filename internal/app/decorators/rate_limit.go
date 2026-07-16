package decorators

import (
	"context"
	"time"

	"github.com/rs/zerolog"

	"github.com/DEEJ4Y/genkitkraft/internal/app/commands"
	"github.com/DEEJ4Y/genkitkraft/internal/app/executors"
	"github.com/DEEJ4Y/genkitkraft/internal/common/errors"
	"github.com/DEEJ4Y/genkitkraft/internal/ports/cache"
)

const (
	rateLimitWindow   = 1 * time.Minute
	rateLimitMaxFails = 5
)

// Compile-time check that RateLimitingLoginDecorator implements the login executor interface.
var _ executors.ExecutorWithReturn[commands.LoginParams, commands.LoginResult] = (*RateLimitingLoginDecorator)(nil)

// RateLimitingLoginDecorator wraps a login executor with per-IP rate limiting:
// at most rateLimitMaxFails failed attempts per rateLimitWindow.
//
// Counting is delegated to the cache port's atomic Increment, so when the cache
// is shared (Redis/Valkey) the limit holds across every instance rather than
// per-process. This is a fixed window — it resets rateLimitWindow after the
// first attempt, not on a rolling basis — which permits a brief burst across a
// window boundary but keeps the sustained rate at rateLimitMaxFails per window.
type RateLimitingLoginDecorator struct {
	inner  executors.ExecutorWithReturn[commands.LoginParams, commands.LoginResult]
	cache  cache.Cache
	logger zerolog.Logger
}

func NewRateLimitingLoginDecorator(
	inner executors.ExecutorWithReturn[commands.LoginParams, commands.LoginResult],
	c cache.Cache,
	logger zerolog.Logger,
) *RateLimitingLoginDecorator {
	return &RateLimitingLoginDecorator{inner: inner, cache: c, logger: logger}
}

func (d *RateLimitingLoginDecorator) Execute(ctx context.Context, params commands.LoginParams) (commands.LoginResult, error) {
	count, err := d.cache.Increment(ctx, params.ClientIP, rateLimitWindow)
	if err != nil {
		// Fail closed. A cache outage already breaks session validation, so
		// admitting unlimited attempts would surrender brute-force protection
		// without making the service usable.
		d.logger.Error().Err(err).Str("ip", params.ClientIP).Msg("rate_limit: cache unavailable, denying login")
		return commands.LoginResult{}, errors.NewAppError(errors.TooManyRequests, "too many login attempts, try again later")
	}
	if count > rateLimitMaxFails {
		return commands.LoginResult{}, errors.NewAppError(errors.TooManyRequests, "too many login attempts, try again later")
	}

	completed := false
	defer func() {
		if !completed {
			// inner.Execute panicked; give back only the attempt we just counted
			// so prior failures for this IP remain counted.
			if err := d.cache.Decrement(ctx, params.ClientIP); err != nil {
				d.logger.Warn().Err(err).Str("ip", params.ClientIP).Msg("rate_limit: releasing attempt after panic failed")
			}
		}
	}()

	result, err := d.inner.Execute(ctx, params)
	completed = true

	if err != nil {
		return result, err
	}

	// Successful login — clear the failure count for this IP.
	if err := d.cache.Delete(ctx, params.ClientIP); err != nil {
		d.logger.Warn().Err(err).Str("ip", params.ClientIP).Msg("rate_limit: clearing count after successful login failed")
	}
	return result, nil
}
