// Package tracemerger combines spans from multiple trace stores into a single
// unified trace store, deduplicating by span ID.
package tracemerger

import (
	"errors"
	"fmt"

	"github.com/user/grpc-tracer/internal/storage"
)

// ErrNoSources is returned when Merge is called with no source stores.
var ErrNoSources = errors.New("tracemerger: at least one source store is required")

// Merger merges spans from multiple TraceStores into a destination store.
type Merger struct {
	dst *storage.TraceStore
}

// New creates a Merger that writes merged results into dst.
func New(dst *storage.TraceStore) (*Merger, error) {
	if dst == nil {
		return nil, errors.New("tracemerger: destination store must not be nil")
	}
	return &Merger{dst: dst}, nil
}

// Merge reads all traces from each source store and adds their spans to the
// destination store. Spans that share a span ID are deduplicated; the first
// occurrence wins.
func (m *Merger) Merge(sources ...*storage.TraceStore) (int, error) {
	if len(sources) == 0 {
		return 0, ErrNoSources
	}

	seen := make(map[string]struct{})
	added := 0

	for i, src := range sources {
		if src == nil {
			return added, fmt.Errorf("tracemerger: source at index %d is nil", i)
		}
		traces := src.GetAllTraces()
		for _, spans := range traces {
			for _, span := range spans {
				if _, dup := seen[span.SpanID]; dup {
					continue
				}
				seen[span.SpanID] = struct{}{}
				m.dst.AddSpan(span)
				added++
			}
		}
	}
	return added, nil
}
