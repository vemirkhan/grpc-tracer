// Package tracerotator provides a rolling-window store that evicts the oldest
// traces once a configurable capacity is reached.
package tracerotator

import (
	"errors"
	"sync"

	"github.com/example/grpc-tracer/internal/storage"
)

// Options controls rotator behaviour.
type Options struct {
	// MaxTraces is the maximum number of distinct trace IDs to retain.
	// When exceeded the oldest trace is evicted. Defaults to 100.
	MaxTraces int
}

func defaultOptions() Options {
	return Options{MaxTraces: 100}
}

// Rotator wraps a TraceStore and evicts the oldest trace when capacity is
// exceeded.
type Rotator struct {
	mu       sync.Mutex
	store    *storage.TraceStore
	order    []string // insertion-ordered trace IDs
	maxTrace int
}

// New creates a Rotator backed by store using opts (nil uses defaults).
func New(store *storage.TraceStore, opts *Options) (*Rotator, error) {
	if store == nil {
		return nil, errors.New("tracerotator: store must not be nil")
	}
	o := defaultOptions()
	if opts != nil {
		if opts.MaxTraces > 0 {
			o.MaxTraces = opts.MaxTraces
		}
	}
	return &Rotator{
		store:    store,
		order:    make([]string, 0, o.MaxTraces),
		maxTrace: o.MaxTraces,
	}, nil
}

// AddSpan records span into the underlying store, evicting the oldest trace
// when the capacity limit is reached.
func (r *Rotator) AddSpan(span storage.Span) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Track insertion order for new trace IDs.
	if !r.known(span.TraceID) {
		if len(r.order) >= r.maxTrace {
			// Evict the oldest trace.
			oldest := r.order[0]
			r.order = r.order[1:]
			r.store.DeleteTrace(oldest)
		}
		r.order = append(r.order, span.TraceID)
	}

	r.store.AddSpan(span)
}

// Store returns the underlying TraceStore for read access.
func (r *Rotator) Store() *storage.TraceStore { return r.store }

// Len returns the number of distinct traces currently held.
func (r *Rotator) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.order)
}

// known reports whether traceID is already tracked (must be called with r.mu held).
func (r *Rotator) known(traceID string) bool {
	for _, id := range r.order {
		if id == traceID {
			return true
		}
	}
	return false
}
