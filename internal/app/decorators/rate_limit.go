package decorators

import (
	"context"
	"sync"
	"time"

	"github.com/DEEJ4Y/genkitkraft/internal/app/commands"
	"github.com/DEEJ4Y/genkitkraft/internal/app/executors"
	"github.com/DEEJ4Y/genkitkraft/internal/common/errors"
)

const (
	rateLimitWindow   = 1 * time.Minute
	rateLimitMaxFails = 5
)

// Compile-time check that RateLimitingLoginDecorator implements the login executor interface.
var _ executors.ExecutorWithReturn[commands.LoginParams, commands.LoginResult] = (*RateLimitingLoginDecorator)(nil)

// RateLimitingLoginDecorator wraps a login executor with per-IP rate limiting.
type RateLimitingLoginDecorator struct {
	inner    executors.ExecutorWithReturn[commands.LoginParams, commands.LoginResult]
	mu       sync.Mutex
	attempts map[string][]time.Time
}

func NewRateLimitingLoginDecorator(
	inner executors.ExecutorWithReturn[commands.LoginParams, commands.LoginResult],
) *RateLimitingLoginDecorator {
	return &RateLimitingLoginDecorator{
		inner:    inner,
		attempts: make(map[string][]time.Time),
	}
}

func (d *RateLimitingLoginDecorator) Execute(ctx context.Context, params commands.LoginParams) (commands.LoginResult, error) {
	if !d.allow(params.ClientIP) {
		return commands.LoginResult{}, errors.NewAppError(errors.TooManyRequests, "too many login attempts, try again later")
	}

	result, err := d.inner.Execute(ctx, params)
	if err != nil {
		d.record(params.ClientIP)
		return result, err
	}

	d.reset(params.ClientIP)
	return result, nil
}

func (d *RateLimitingLoginDecorator) allow(ip string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.prune(ip)
	return len(d.attempts[ip]) < rateLimitMaxFails
}

func (d *RateLimitingLoginDecorator) record(ip string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.attempts[ip] = append(d.attempts[ip], time.Now())
}

func (d *RateLimitingLoginDecorator) reset(ip string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.attempts, ip)
}

func (d *RateLimitingLoginDecorator) prune(ip string) {
	cutoff := time.Now().Add(-rateLimitWindow)
	attempts := d.attempts[ip]
	i := 0
	for _, t := range attempts {
		if t.After(cutoff) {
			attempts[i] = t
			i++
		}
	}
	d.attempts[ip] = attempts[:i]
}
