// Package tagselector provides utilities for matching and selecting spans
// based on arbitrary key-value tag criteria.
package tagselector

import "github.com/your-org/grpc-tracer/internal/storage"

// Criteria holds the set of tags a span must match.
type Criteria struct {
	// Tags is a map of key→value pairs that must all be present in a span's
	// metadata for it to be selected. An empty map matches every span.
	Tags map[string]string
}

// MatchSpan reports whether the given span satisfies all tag criteria.
func MatchSpan(span storage.Span, c Criteria) bool {
	if len(c.Tags) == 0 {
		return true
	}
	, want := range c.Tags {
		got, ok := span.Metadata[k]
		if !ok || got != want {
			return false
		}
	}
	return true
}

// SelectSpans returns only the spans from the provided slice that satisfy
// all tag criteria.
func SelectSpans(spans []storage.Span, c Criteria) []storage.Span {
	var out []storage.Span
	for _, s := range spans {
		if MatchSpan(s, c) {
			out = append(out, s)
		}
	}
	return out
}

// SelectTraces filters a map of traceID→spans, returning only traces that
// contain at least one span matching the criteria.
func SelectTraces(traces map[string][]storage.Span, c Criteria) map[string][]storage.Span {
	out := make(map[string][]storage.Span)
	for id, spans := range traces {
		matched := SelectSpans(spans, c)
		if len(matched) > 0 {
			out[id] = matched
		}
	}
	return out
}
