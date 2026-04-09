// Package filter provides utilities for filtering traces and spans
// based on various criteria such as service name, status, or duration.
package filter

import (
	"time"

	"github.com/user/grpc-tracer/internal/storage"
)

// Criteria holds the filtering parameters for trace queries.
type Criteria struct {
	ServiceName string
	MinDuration time.Duration
	MaxDuration time.Duration
	OnlyErrors  bool
	TraceID     string
}

// Filter applies the given Criteria to a slice of spans and returns
// only those spans that satisfy all specified conditions.
func Filter(spans []storage.Span, c Criteria) []storage.Span {
	var result []storage.Span
	for _, s := range spans {
		if c.ServiceName != "" && s.ServiceName != c.ServiceName {
			continue
		}
		if c.TraceID != "" && s.TraceID != c.TraceID {
			continue
		}
		dur := s.EndTime.Sub(s.StartTime)
		if c.MinDuration > 0 && dur < c.MinDuration {
			continue
		}
		if c.MaxDuration > 0 && dur > c.MaxDuration {
			continue
		}
		if c.OnlyErrors && s.Error == "" {
			continue
		}
		result = append(result, s)
	}
	return result
}

// FilterTraces applies Criteria across all traces in the store and returns
// a map of traceID -> matching spans.
func FilterTraces(store *storage.TraceStore, c Criteria) map[string][]storage.Span {
	all := store.GetAllTraces()
	out := make(map[string][]storage.Span)
	for traceID, spans := range all {
		filtered := Filter(spans, c)
		if len(filtered) > 0 {
			out[traceID] = filtered
		}
	}
	return out
}
