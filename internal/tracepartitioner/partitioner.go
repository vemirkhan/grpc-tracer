// Package tracepartitioner splits spans from a trace store into
// named partitions based on a user-supplied key function.
package tracepartitioner

import (
	"errors"
	"sync"

	"github.com/example/grpc-tracer/internal/storage"
)

// KeyFunc derives a partition key from a span.
type KeyFunc func(span storage.Span) string

// Partitioner distributes spans across named sub-stores.
type Partitioner struct {
	mu      sync.RWMutex
	source  *storage.TraceStore
	keyFn   KeyFunc
	parts   map[string]*storage.TraceStore
}

// New creates a Partitioner that reads from source and routes spans
// via keyFn. Returns an error when either argument is nil.
func New(source *storage.TraceStore, keyFn KeyFunc) (*Partitioner, error) {
	if source == nil {
		return nil, errors.New("tracepartitioner: source store must not be nil")
	}
	if keyFn == nil {
		return nil, errors.New("tracepartitioner: key function must not be nil")
	}
	return &Partitioner{
		source: source,
		keyFn:  keyFn,
		parts:  make(map[string]*storage.TraceStore),
	}, nil
}

// Partition reads all traces from the source store, routes each span
// into the appropriate sub-store keyed by keyFn, and returns the
// resulting partition map.
func (p *Partitioner) Partition() map[string]*storage.TraceStore {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Reset partitions on every call so results are deterministic.
	p.parts = make(map[string]*storage.TraceStore)

	for _, trace := range p.source.GetAllTraces() {
		for _, span := range trace {
			key := p.keyFn(span)
			if key == "" {
				key = "__default__"
			}
			if _, ok := p.parts[key]; !ok {
				p.parts[key] = storage.NewTraceStore()
			}
			p.parts[key].AddSpan(span)
		}
	}

	// Return a shallow copy so callers cannot mutate internal state.
	out := make(map[string]*storage.TraceStore, len(p.parts))
	for k, v := range p.parts {
		out[k] = v
	}
	return out
}

// Keys returns the partition keys produced by the last Partition call.
func (p *Partitioner) Keys() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	keys := make([]string, 0, len(p.parts))
	for k := range p.parts {
		keys = append(keys, k)
	}
	return keys
}

// ByService is a built-in KeyFunc that partitions by service name.
func ByService(span storage.Span) string { return span.Service }

// ByMethod is a built-in KeyFunc that partitions by RPC method.
func ByMethod(span storage.Span) string { return span.Method }
