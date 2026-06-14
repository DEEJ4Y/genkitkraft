package decorators

import (
	"context"
	"encoding/json"
	"sync"
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
type RateLimitingLoginDecorator struct {
	inner executors.ExecutorWithReturn[commands.LoginParams, commands.LoginResult]
	cache cache.Cache
	mu    sync.Mutex
}

func NewRateLimitingLoginDecorator(
	inner executors.ExecutorWithReturn[commands.LoginParams, commands.LoginResult],
	c cache.Cache,
) *RateLimitingLoginDecorator {
	return &RateLimitingLoginDecorator{inner: inner, cache: c}
}

func (d *RateLimitingLoginDecorator) Execute(ctx context.Context, params commands.LoginParams) (commands.LoginResult, error) {
	if !d.checkAndRecord(ctx, params.ClientIP) {
		return commands.LoginResult{}, errors.NewAppError(errors.TooManyRequests, "too many login attempts, try again later")
	}

	result, err := d.inner.Execute(ctx, params)
	if err != nil {
		return result, err
	}

	// Successful login — clear the failure count for this IP.
	d.mu.Lock()
	_ = d.cache.Delete(ctx, params.ClientIP)
	d.mu.Unlock()
	return result, nil
}

// checkAndRecord atomically checks whether the IP is under the rate limit and,
// if so, pre-records the attempt. Returns false if the limit is already reached.
// The lock is not held during inner.Execute so slow operations (e.g. bcrypt) do
// not block concurrent requests from other IPs.
func (d *RateLimitingLoginDecorator) checkAndRecord(ctx context.Context, ip string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	attempts := d.getAttempts(ctx, ip)
	attempts = prune(attempts)
	if len(attempts) >= rateLimitMaxFails {
		return false
	}
	attempts = append(attempts, time.Now().UnixNano())
	d.setAttempts(ctx, ip, attempts)
	return true
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
