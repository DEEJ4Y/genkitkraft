package streamregistry

import "context"

// Registry tracks the cancel functions for in-flight, detached-context stream
// generations on this process, keyed by message ID. It exists so an explicit
// "stop generation" request can cancel a goroutine that is deliberately no
// longer tied to the HTTP request that started it.
//
// Registration is best-effort and process-local: a message started on a
// different instance (or one whose generation already finished) simply won't
// be found, and Cancel reports that back rather than erroring.
type Registry interface {
	// Register associates messageID with cancel, so a later Cancel(messageID)
	// can stop the generation. Overwrites any prior registration for the same ID.
	Register(messageID string, cancel context.CancelFunc)
	// Cancel calls the registered cancel function for messageID, if any, and
	// removes it. Reports whether a registration was found.
	Cancel(messageID string) bool
	// Unregister removes messageID without invoking its cancel function. Called
	// once a stream finishes on its own so the registry doesn't grow unbounded.
	Unregister(messageID string)
}
