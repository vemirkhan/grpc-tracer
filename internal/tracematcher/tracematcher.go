// Package tracematcher provides pattern-based matching for traces and spans,
// allowing callers to find traces that satisfy a set of conditions.
package tracematcher

import (
	"strings"
	"time"

	"github.com/user/grpc-tracer/internal/storage"
)

// Criteria holds the conditions a trace must satisfy to be considered a match.
type Criteria struct {
	// TraceID matches the exact trace ID (empty = any).
	TraceID string
	// ServiceName matches spans whose ServiceName contains this value (case-insensitive, empty = any).
	ServiceName string
	// Method matches spans whose Method contains this value (case-insensitive, empty = any).
	Method string
	// MinDuration filters out traces whose total duration is below this threshold.
	MinDuration time.Duration
	// OnlyErrors keeps only traces that contain at least one span with a non-empty Error field.
	OnlyErrors bool
	// TagKey / TagValue filter traces that have at least one span containing the given tag pair.
	TagKey   string
	TagValue string
}

// Match returns true when the given trace satisfies all non-zero fields in c.
func Match(trace storage.Trace, c Criteria) bool {
	if c.TraceID != "" && trace.TraceID != c.TraceID {
		return false
	}

	var (
		serviceOK  = c.ServiceName == ""
		methodOK   = c.Method == ""
		errorFound = !c.OnlyErrors
		tagFound   = c.TagKey == ""
		earliestStart time.Time
		latestEnd     time.Time
	)

	for i, span := range trace.Spans {
		if !c.ServiceName == false && strings.Contains(
			strings.ToLower(span.ServiceName), strings.ToLower(c.ServiceName)) {
			serviceOK = true
		}
		if !methodOK && strings.Contains(
			strings.ToLower(span.Method), strings.ToLower(c.Method)) {
			methodOK = true
		}
		if c.OnlyErrors && span.Error != "" {
			errorFound = true
		}
		if c.TagKey != "" {
			if v, ok := span.Tags[c.TagKey]; ok && (c.TagValue == "" || v == c.TagValue) {
				tagFound = true
			}
		}
		if i == 0 || span.StartTime.Before(earliestStart) {
			earliestStart = span.StartTime
		}
		end := span.StartTime.Add(span.Duration)
		if end.After(latestEnd) {
			latestEnd = end
		}
	}

	if c.MinDuration > 0 {
		if latestEnd.Sub(earliestStart) < c.MinDuration {
			return false
		}
	}

	return serviceOK && methodOK && errorFound && tagFound
}

// MatchAll filters store for traces that satisfy c and returns them.
func MatchAll(store *storage.TraceStore, c Criteria) []storage.Trace {
	all := store.GetAllTraces()
	out := make([]storage.Trace, 0, len(all))
	for _, t := range all {
		if Match(t, c) {
			out = append(out, t)
		}
	}
	return out
}
