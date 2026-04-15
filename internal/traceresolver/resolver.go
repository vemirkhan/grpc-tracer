// Package traceresolver resolves a span's full ancestry chain by walking
// parent references stored in the trace store.
package traceresolver

import (
	"errors"
	"fmt"

	"github.com/user/grpc-tracer/internal/storage"
)

// ErrNotFound is returned when a trace or span cannot be located.
var ErrNotFound = errors.New("traceresolver: not found")

// Resolver walks a trace and resolves the ancestor chain for any given span.
type Resolver struct {
	store *storage.TraceStore
}

// New creates a Resolver backed by the provided TraceStore.
// Returns an error if store is nil.
func New(store *storage.TraceStore) (*Resolver, error) {
	if store == nil {
		return nil, errors.New("traceresolver: store must not be nil")
	}
	return &Resolver{store: store}, nil
}

// Ancestors returns the ordered list of spans from the root down to (but not
// including) the span identified by spanID. The first element is always the
// root span; the last is the direct parent of spanID.
//
// Returns ErrNotFound if the trace or span does not exist.
func (r *Resolver) Ancestors(traceID, spanID string) ([]storage.Span, error) {
	spans, err := r.store.GetTrace(traceID)
	if err != nil || len(spans) == 0 {
		return nil, fmt.Errorf("%w: trace %q", ErrNotFound, traceID)
	}

	// Index spans by ID for O(1) lookup.
	byID := make(map[string]storage.Span, len(spans))
	for _, s := range spans {
		byID[s.SpanID] = s
	}

	target, ok := byID[spanID]
	if !ok {
		return nil, fmt.Errorf("%w: span %q in trace %q", ErrNotFound, spanID, traceID)
	}

	// Walk parent pointers to build the chain.
	var chain []storage.Span
	current := target
	visited := make(map[string]struct{})

	for current.ParentSpanID != "" {
		if _, seen := visited[current.SpanID]; seen {
			break // guard against cycles
		}
		visited[current.SpanID] = struct{}{}

		parent, exists := byID[current.ParentSpanID]
		if !exists {
			break
		}
		chain = append(chain, parent)
		current = parent
	}

	// Reverse so the chain runs root → direct parent.
	for i, j := 0, len(chain)-1; i < j; i, j = i+1, j-1 {
		chain[i], chain[j] = chain[j], chain[i]
	}
	return chain, nil
}

// Root returns the root span of the given trace (the span with no parent).
// Returns ErrNotFound if the trace is empty or no root span exists.
func (r *Resolver) Root(traceID string) (storage.Span, error) {
	spans, err := r.store.GetTrace(traceID)
	if err != nil || len(spans) == 0 {
		return storage.Span{}, fmt.Errorf("%w: trace %q", ErrNotFound, traceID)
	}
	for _, s := range spans {
		if s.ParentSpanID == "" {
			return s, nil
		}
	}
	return storage.Span{}, fmt.Errorf("%w: no root span in trace %q", ErrNotFound, traceID)
}
