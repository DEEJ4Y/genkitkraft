package decorators

import (
	"context"
	"encoding/json"
	"time"

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

// RateLimitingLoginDecorator wraps a login executor with per-IP rate limiting.
// Attempt counts are persisted via the cache port. Read-modify-write is not atomic;
// a distributed cache backend would need atomic ops for strict correctness.
type RateLimitingLoginDecorator struct {
	inner executors.ExecutorWithReturn[commands.LoginParams, commands.LoginResult]
	cache cache.Cache
}

func NewRateLimitingLoginDecorator(
	inner executors.ExecutorWithReturn[commands.LoginParams, commands.LoginResult],
	c cache.Cache,
) *RateLimitingLoginDecorator {
	return &RateLimitingLoginDecorator{inner: inner, cache: c}
}

func (d *RateLimitingLoginDecorator) Execute(ctx context.Context, params commands.LoginParams) (commands.LoginResult, error) {
	if !d.allow(ctx, params.ClientIP) {
		return commands.LoginResult{}, errors.NewAppError(errors.TooManyRequests, "too many login attempts, try again later")
	}

	result, err := d.inner.Execute(ctx, params)
	if err != nil {
		d.record(ctx, params.ClientIP)
		return result, err
	}

	_ = d.cache.Delete(ctx, params.ClientIP)
	return result, nil
}

func (d *RateLimitingLoginDecorator) allow(ctx context.Context, ip string) bool {
	attempts := d.getAttempts(ctx, ip)
	attempts = prune(attempts)
	return len(attempts) < rateLimitMaxFails
}

func (d *RateLimitingLoginDecorator) record(ctx context.Context, ip string) {
	attempts := d.getAttempts(ctx, ip)
	attempts = prune(attempts)
	attempts = append(attempts, time.Now().UnixNano())
	d.setAttempts(ctx, ip, attempts)
}

func (d *RateLimitingLoginDecorator) getAttempts(ctx context.Context, ip string) []int64 {
	val, ok := d.cache.Get(ctx, ip)
	if !ok {
		return nil
	}
	var attempts []int64
	if err := json.Unmarshal([]byte(val), &attempts); err != nil {
		return nil
	}
	return attempts
}

func (d *RateLimitingLoginDecorator) setAttempts(ctx context.Context, ip string, attempts []int64) {
	data, err := json.Marshal(attempts)
	if err != nil {
		return
	}
	_ = d.cache.Set(ctx, ip, string(data), rateLimitWindow)
}

// prune removes attempt timestamps that fall outside the rolling window.
func prune(attempts []int64) []int64 {
	cutoff := time.Now().Add(-rateLimitWindow).UnixNano()
	i := 0
	for _, t := range attempts {
		if t > cutoff {
			attempts[i] = t
			i++
		}
	}
	return attempts[:i]
}
