// Package cachestreamregistry implements the stream registry port on top of the
// cache port, so cross-instance "stop generation" signalling is governed by the
// same CACHE_PROVIDER as sessions and login rate limiting.
package cachestreamregistry

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog"

	"github.com/DEEJ4Y/genkitkraft/internal/ports/cache"
	streamregistry "github.com/DEEJ4Y/genkitkraft/internal/ports/stream_registry"
)

var _ streamregistry.Registry = (*Registry)(nil)

const (
	// DefaultPollInterval is the interval used in production wiring.
	DefaultPollInterval = 1 * time.Second
	// cacheOpTimeout bounds every Get/Set/Delete this adapter makes against the
	// cache, so a slow or partitioned backend cannot hang a caller — most notably
	// Cancel, which runs synchronously on the HTTP request path.
	cacheOpTimeout = 5 * time.Second
	// minSignalTTL is a floor under signalTTLFor, so a very short poll interval
	// (as in tests) cannot shrink the signal's TTL below a value that would make
	// it expire mid-flight.
	minSignalTTL = 5 * time.Second
	// cancelSignalValue is written to mark a pending remote cancel. Its presence
	// is the signal; the content is never read.
	cancelSignalValue = "1"
)

// Registry tracks in-flight stream cancel functions locally — a
// context.CancelFunc is an in-process closure and can never be moved to a
// remote cache — while also relaying cancel requests through the cache port so
// an instance different from the one that received the request can still stop
// the generation it owns. Each registration polls the cache for a pending
// signal at pollInterval; under CACHE_PROVIDER=memory this poll reduces to an
// in-process map lookup, so single-instance deployments pay no network cost.
type Registry struct {
	mu      sync.Mutex
	cancels map[string]context.CancelFunc

	cache        cache.Cache
	pollInterval time.Duration
	signalTTL    time.Duration
	logger       zerolog.Logger
}

// NewRegistry creates a Registry backed by c, polling for remote cancel
// signals every pollInterval.
func NewRegistry(c cache.Cache, pollInterval time.Duration, logger zerolog.Logger) *Registry {
	return &Registry{
		cancels:      make(map[string]context.CancelFunc),
		cache:        c,
		pollInterval: pollInterval,
		signalTTL:    signalTTLFor(pollInterval),
		logger:       logger,
	}
}

// signalTTLFor bounds the cancel signal's lifetime to comfortably outlive one
// poll tick. It is a transient "please stop" flag, not a durable instruction,
// so it stays far below a generation's overall timeout.
func signalTTLFor(pollInterval time.Duration) time.Duration {
	if ttl := pollInterval * 10; ttl > minSignalTTL {
		return ttl
	}
	return minSignalTTL
}

func (r *Registry) Register(messageID string, cancel context.CancelFunc) {
	r.mu.Lock()
	r.cancels[messageID] = cancel
	r.mu.Unlock()

	go r.poll(messageID)
}

// poll watches for a remote cancel signal for messageID until the
// registration is gone — locally cancelled, unregistered, or (once this
// instance notices the signal itself) remotely cancelled.
func (r *Registry) poll(messageID string) {
	ticker := time.NewTicker(r.pollInterval)
	defer ticker.Stop()

	for range ticker.C {
		if !r.isRegistered(messageID) {
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), cacheOpTimeout)
		_, found, err := r.cache.Get(ctx, messageID)
		cancel()
		if err != nil {
			r.logger.Warn().Err(err).Str("message_id", messageID).
				Msg("stream_registry: checking for a remote cancel signal failed")
			continue
		}
		if found {
			r.Cancel(messageID)
			return
		}
	}
}

func (r *Registry) isRegistered(messageID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, found := r.cancels[messageID]
	return found
}

func (r *Registry) Cancel(messageID string) bool {
	r.mu.Lock()
	cancel, found := r.cancels[messageID]
	delete(r.cancels, messageID)
	r.mu.Unlock()

	// Written unconditionally, regardless of found: this instance cannot tell
	// "no one owns this stream" apart from "some other instance owns it", and a
	// signal is harmless in the former case since nothing is polling for it.
	ctx, cancelTimeout := context.WithTimeout(context.Background(), cacheOpTimeout)
	err := r.cache.Set(ctx, messageID, cancelSignalValue, r.signalTTL)
	cancelTimeout()
	if err != nil {
		r.logger.Warn().Err(err).Str("message_id", messageID).
			Msg("stream_registry: signalling a remote cancel failed")
	}

	if !found {
		return false
	}
	cancel()
	return true
}

func (r *Registry) Unregister(messageID string) {
	r.mu.Lock()
	delete(r.cancels, messageID)
	r.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), cacheOpTimeout)
	defer cancel()
	if err := r.cache.Delete(ctx, messageID); err != nil {
		r.logger.Warn().Err(err).Str("message_id", messageID).
			Msg("stream_registry: clearing a cancel signal failed")
	}
}
