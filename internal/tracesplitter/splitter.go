// Package tracesplitter provides functionality to split a single trace into
// multiple sub-traces based on a user-supplied predicate. This is useful when
// a long-running trace contains logically independent segments that should be
// analysed or exported separately.
package tracesplitter

import (
	"errors"
	"fmt"

	"github.com/your-org/grpc-tracer/internal/storage"
)

// Predicate decides which "bucket" a span belongs to. It receives the span and
// returns a non-empty string key that identifies the target sub-trace. Spans
// that return an empty key are dropped.
type Predicate func(span storage.Span) string

// Splitter splits spans from a source trace into destination stores.
type Splitter struct {
	src  *storage.TraceStore
	dsts map[string]*storage.TraceStore
	pred Predicate
}

// New creates a Splitter. src must not be nil and pred must not be nil.
func New(src *storage.TraceStore, pred Predicate) (*Splitter, error) {
	if src == nil {
		return nil, errors.New("tracesplitter: source store must not be nil")
	}
	if pred == nil {
		return nil, errors.New("tracesplitter: predicate must not be nil")
	}
	return &Splitter{
		src:  src,
		dsts: make(map[string]*storage.TraceStore),
		pred: pred,
	}, nil
}

// Split reads every span belonging to traceID from the source store and
// distributes them into per-bucket destination stores. It returns a map of
// bucket-key → TraceStore containing the resulting sub-traces.
func (s *Splitter) Split(traceID string) (map[string]*storage.TraceStore, error) {
	spans, ok := s.src.GetTrace(traceID)
	if !ok {
		return nil, fmt.Errorf("tracesplitter: trace %q not found", traceID)
	}

	buckets := make(map[string]*storage.TraceStore)
	for _, sp := range spans {
		key := s.pred(sp)
		if key == "" {
			continue
		}
		if _, exists := buckets[key]; !exists {
			buckets[key] = storage.NewTraceStore()
		}
		buckets[key].AddSpan(sp)
	}
	return buckets, nil
}

// ByService is a ready-made Predicate that buckets spans by their ServiceName.
func ByService(span storage.Span) string {
	return span.ServiceName
}

// ByTag returns a Predicate that buckets spans by the value of the given tag
// key. Spans that lack the tag are dropped (empty string returned).
func ByTag(key string) Predicate {
	return func(span storage.Span) string {
		if span.Tags == nil {
			return ""
		}
		return span.Tags[key]
	}
}
