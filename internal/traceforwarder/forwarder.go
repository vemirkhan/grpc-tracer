// Package traceforwarder forwards spans from a source store to one or more
// destination stores, optionally filtering by a predicate before forwarding.
package traceforwarder

import (
	"errors"
	"sync"

	"github.com/example/grpc-tracer/internal/storage"
)

// Predicate is a function that returns true if a span should be forwarded.
type Predicate func(span storage.Span) bool

// Forwarder copies spans from a source store to destination stores.
type Forwarder struct {
	mu           sync.Mutex
	source       *storage.TraceStore
	destinations []*storage.TraceStore
	predicate    Predicate
}

// New creates a new Forwarder. source must not be nil.
// If predicate is nil, all spans are forwarded.
func New(source *storage.TraceStore, predicate Predicate, destinations ...*storage.TraceStore) (*Forwarder, error) {
	if source == nil {
		return nil, errors.New("traceforwarder: source store must not be nil")
	}
	for i, d := range destinations {
		if d == nil {
			return nil, fmt.Errorf("traceforwarder: destination at index %d must not be nil", i)
		}
	}
	if predicate == nil {
		predicate = func(storage.Span) bool { return true }
	}
	return &Forwarder{
		source:       source,
		destinations: destinations,
		predicate:    predicate,
	}, nil
}

// AddDestination appends a destination store at runtime.
func (f *Forwarder) AddDestination(dest *storage.TraceStore) error {
	if dest == nil {
		return errors.New("traceforwarder: destination must not be nil")
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.destinations = append(f.destinations, dest)
	return nil
}

// ForwardTrace reads all spans for traceID from the source and writes those
// accepted by the predicate to every destination store.
func (f *Forwarder) ForwardTrace(traceID string) (int, error) {
	spans, ok := f.source.GetTrace(traceID)
	if !ok {
		return 0, fmt.Errorf("traceforwarder: trace %q not found", traceID)
	}

	f.mu.Lock()
	dests := make([]*storage.TraceStore, len(f.destinations))
	copy(dests, f.destinations)
	f.mu.Unlock()

	forwarded := 0
	for _, span := range spans {
		if !f.predicate(span) {
			continue
		}
		for _, dest := range dests {
			dest.AddSpan(span)
		}
		forwarded++
	}
	return forwarded, nil
}

// ForwardAll forwards every trace in the source store.
func (f *Forwarder) ForwardAll() (int, error) {
	traces := f.source.GetAllTraces()
	total := 0
	for traceID := range traces {
		n, err := f.ForwardTrace(traceID)
		if err != nil {
			return total, err
		}
		total += n
	}
	return total, nil
}
