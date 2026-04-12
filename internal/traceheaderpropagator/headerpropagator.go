// Package traceheaderpropagator provides HTTP header-based trace context propagation,
// allowing trace IDs and span IDs to be forwarded via standard HTTP headers
// such as X-Trace-Id, X-Span-Id, and X-Parent-Span-Id.
package traceheaderpropagator

import (
	"net/http"
	"strings"
)

const (
	HeaderTraceID    = "x-trace-id"
	HeaderSpanID     = "x-span-id"
	HeaderParentSpan = "x-parent-span-id"
	HeaderSampled    = "x-sampled"
)

// TraceHeaders holds the propagated trace context extracted from HTTP headers.
type TraceHeaders struct {
	TraceID      string
	SpanID       string
	ParentSpanID string
	Sampled      bool
}

// Inject writes the given TraceHeaders into the provided http.Header map.
// Empty fields are skipped.
func Inject(h http.Header, t TraceHeaders) {
	if t.TraceID != "" {
		h.Set(HeaderTraceID, t.TraceID)
	}
	if t.SpanID != "" {
		h.Set(HeaderSpanID, t.SpanID)
	}
	if t.ParentSpanID != "" {
		h.Set(HeaderParentSpan, t.ParentSpanID)
	}
	if t.Sampled {
		h.Set(HeaderSampled, "1")
	} else {
		h.Set(HeaderSampled, "0")
	}
}

// Extract reads trace context from the provided http.Header and returns
// a TraceHeaders struct. Missing headers result in zero-value fields.
func Extract(h http.Header) TraceHeaders {
	return TraceHeaders{
		TraceID:      strings.TrimSpace(h.Get(HeaderTraceID)),
		SpanID:       strings.TrimSpace(h.Get(HeaderSpanID)),
		ParentSpanID: strings.TrimSpace(h.Get(HeaderParentSpan)),
		Sampled:      strings.TrimSpace(h.Get(HeaderSampled)) == "1",
	}
}

// IsValid returns true if the TraceHeaders contains at minimum a non-empty TraceID.
func (t TraceHeaders) IsValid() bool {
	return t.TraceID != ""
}
