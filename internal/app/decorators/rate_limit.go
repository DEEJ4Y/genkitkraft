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
	ts, ok := d.checkAndRecord(ctx, params.ClientIP)
	if !ok {
		return commands.LoginResult{}, errors.NewAppError(errors.TooManyRequests, "too many login attempts, try again later")
	}

	completed := false
	defer func() {
		if !completed {
			// inner.Execute panicked; remove only the slot we pre-recorded so prior
			// failures for this IP remain counted.
			d.mu.Lock()
			d.removeAttempt(ctx, params.ClientIP, ts)
			d.mu.Unlock()
		}
	}()

	result, err := d.inner.Execute(ctx, params)
	completed = true

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
// if so, pre-records the attempt. Returns the recorded timestamp and true if
// allowed, or zero and false if the limit is already reached.
// The lock is not held during inner.Execute so slow operations (e.g. bcrypt) do
// not block concurrent requests from other IPs.
func (d *RateLimitingLoginDecorator) checkAndRecord(ctx context.Context, ip string) (int64, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()

	attempts := d.getAttempts(ctx, ip)
	attempts = prune(attempts)
	if len(attempts) >= rateLimitMaxFails {
		return 0, false
	}
	ts := time.Now().UnixNano()
	attempts = append(attempts, ts)
	d.setAttempts(ctx, ip, attempts)
	return ts, true
}

// removeAttempt removes the single entry with timestamp ts for ip. Called on
// panic to undo the pre-recorded slot without erasing prior failures.
func (d *RateLimitingLoginDecorator) removeAttempt(ctx context.Context, ip string, ts int64) {
	attempts := d.getAttempts(ctx, ip)
	out := attempts[:0]
	removed := false
	for _, t := range attempts {
		if !removed && t == ts {
			removed = true
			continue
		}
		out = append(out, t)
	}
	if len(out) == 0 {
		_ = d.cache.Delete(ctx, ip)
	} else {
		d.setAttempts(ctx, ip, out)
	}
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
