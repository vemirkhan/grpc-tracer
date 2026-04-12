// Package tracededuplicator provides span-level deduplication for trace stores.
// Duplicate spans are identified by their SpanID; only the first occurrence is kept.
package tracededuplicator

import (
	"errors"
	"sync"

	"github.com/user/grpc-tracer/internal/storage"
)

// Deduplicator filters duplicate spans before they reach a destination store.
type Deduplicator struct {
	mu   sync.Mutex
	seen map[string]struct{}
	dest *storage.TraceStore
}

// New creates a Deduplicator that writes unique spans to dest.
// Returns an error if dest is nil.
func New(dest *storage.TraceStore) (*Deduplicator, error) {
	if dest == nil {
		return nil, errors.New("tracededuplicator: destination store must not be nil")
	}
	return &Deduplicator{
		seen: make(map[string]struct{}),
		dest: dest,
	}, nil
}

// AddSpan records the span in the destination store only if its SpanID has not
// been seen before. Returns true if the span was accepted, false if it was a
// duplicate.
func (d *Deduplicator) AddSpan(span storage.Span) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, exists := d.seen[span.SpanID]; exists {
		return false
	}
	d.seen[span.SpanID] = struct{}{}
	d.dest.AddSpan(span)
	return true
}

// Reset clears the set of seen span IDs without modifying the destination store.
func (d *Deduplicator) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.seen = make(map[string]struct{})
}

// SeenCount returns the number of unique span IDs recorded so far.
func (d *Deduplicator) SeenCount() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.seen)
}
