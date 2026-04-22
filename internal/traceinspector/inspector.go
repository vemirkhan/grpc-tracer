// Package traceinspector provides span-level inspection utilities,
// allowing callers to query structural properties of a trace such as
// critical-path detection and longest-span identification.
package traceinspector

import (
	"errors"
	"sort"
	"time"

	"github.com/example/grpc-tracer/internal/storage"
)

// ErrTraceNotFound is returned when the requested trace does not exist.
var ErrTraceNotFound = errors.New("traceinspector: trace not found")

// SpanSummary holds lightweight inspection results for a single span.
type SpanSummary struct {
	SpanID   string
	Service  string
	Method   string
	Duration time.Duration
	HasError bool
}

// Inspector analyses traces stored in a TraceStore.
type Inspector struct {
	store *storage.TraceStore
}

// New creates an Inspector backed by the provided store.
// It returns an error if store is nil.
func New(store *storage.TraceStore) (*Inspector, error) {
	if store == nil {
		return nil, errors.New("traceinspector: store must not be nil")
	}
	return &Inspector{store: store}, nil
}

// LongestSpan returns a summary of the span with the greatest duration
// within the given trace. It returns ErrTraceNotFound when the trace is
// absent or contains no spans.
func (i *Inspector) LongestSpan(traceID string) (SpanSummary, error) {
	spans := i.store.GetTrace(traceID)
	if len(spans) == 0 {
		return SpanSummary{}, ErrTraceNotFound
	}

	best := spans[0]
	for _, s := range spans[1:] {
		if s.Duration > best.Duration {
			best = s
		}
	}
	return toSummary(best), nil
}

// CriticalPath returns span summaries ordered from longest to shortest
// duration, representing the dominant execution path through the trace.
func (i *Inspector) CriticalPath(traceID string) ([]SpanSummary, error) {
	spans := i.store.GetTrace(traceID)
	if len(spans) == 0 {
		return nil, ErrTraceNotFound
	}

	sorted := make([]storage.Span, len(spans))
	copy(sorted, spans)
	sort.Slice(sorted, func(a, b int) bool {
		return sorted[a].Duration > sorted[b].Duration
	})

	out := make([]SpanSummary, len(sorted))
	for idx, s := range sorted {
		out[idx] = toSummary(s)
	}
	return out, nil
}

func toSummary(s storage.Span) SpanSummary {
	return SpanSummary{
		SpanID:   s.SpanID,
		Service:  s.ServiceName,
		Method:   s.Method,
		Duration: s.Duration,
		HasError: s.Error != "",
	}
}
