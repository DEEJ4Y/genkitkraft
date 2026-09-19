package streamregistry

import "context"

// Registry tracks the cancel functions for in-flight, detached-context stream
// generations, keyed by message ID. It exists so an explicit "stop generation"
// request can cancel a goroutine that is deliberately no longer tied to the
// HTTP request that started it.
//
// A context.CancelFunc is an in-process closure, so invoking it always happens
// on the instance that owns the registration — that part of Cancel's behavior
// is the same for every adapter. Whether a Cancel call made on a *different*
// instance can still reach that registration is adapter-specific: it requires
// the adapter to relay the request through a store shared across instances.
// Register and Cancel report only what this instance observed, so "no
// registration found" covers a generation that already finished, one that
// never started, and (for an adapter with no cross-instance reach) one owned
// by another instance — all as a normal race, not an error.
type Registry interface {
	// Register associates messageID with cancel, so a later Cancel(messageID)
	// can stop the generation. Overwrites any prior registration for the same ID.
	Register(messageID string, cancel context.CancelFunc)
	// Cancel calls the registered cancel function for messageID, if any, and
	// removes it. Reports whether a registration was found on this instance.
	Cancel(messageID string) bool
	// Unregister removes messageID without invoking its cancel function. Called
	// once a stream finishes on its own so the registry doesn't grow unbounded.
	Unregister(messageID string)
}
