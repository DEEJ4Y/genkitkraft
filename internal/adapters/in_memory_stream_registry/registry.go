package inmemorystreamregistry

import (
	"context"
	"sync"

	streamregistry "github.com/DEEJ4Y/genkitkraft/internal/ports/stream_registry"
)

var _ streamregistry.Registry = (*Registry)(nil)

// Registry is a process-local, mutex-guarded map of in-flight stream cancel
// functions. Like in_memory_cache, it only coordinates within a single
// instance — a multi-instance deployment only ever cancels a stream on the
// instance that started it.
type Registry struct {
	mu      sync.Mutex
	cancels map[string]context.CancelFunc
}

// NewRegistry creates an empty in-memory stream registry.
func NewRegistry() *Registry {
	return &Registry{cancels: make(map[string]context.CancelFunc)}
}

func (r *Registry) Register(messageID string, cancel context.CancelFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cancels[messageID] = cancel
}

func (r *Registry) Cancel(messageID string) bool {
	r.mu.Lock()
	cancel, found := r.cancels[messageID]
	delete(r.cancels, messageID)
	r.mu.Unlock()

	if !found {
		return false
	}
	cancel()
	return true
}

func (r *Registry) Unregister(messageID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.cancels, messageID)
}
