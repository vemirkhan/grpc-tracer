// Package redactor provides utilities for scrubbing sensitive metadata
// keys from gRPC span data before storage or export.
package redactor

import (
	"strings"

	"github.com/user/grpc-tracer/internal/storage"
)

// DefaultSensitiveKeys contains common metadata keys that should be redacted.
var DefaultSensitiveKeys = []string{
	"authorization",
	"x-api-key",
	"cookie",
	"set-cookie",
	"x-auth-token",
	"password",
}

// Redactor scrubs sensitive keys from span metadata.
type Redactor struct {
	sensitiveKeys map[string]struct{}
	placeholder   string
}

// New creates a Redactor with the given sensitive keys.
// If keys is empty, DefaultSensitiveKeys is used.
func New(keys []string, placeholder string) *Redactor {
	if len(keys) == 0 {
		keys = DefaultSensitiveKeys
	}
	if placeholder == "" {
		placeholder = "[REDACTED]"
	}
	km := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		km[strings.ToLower(k)] = struct{}{}
	}
	return &Redactor{sensitiveKeys: km, placeholder: placeholder}
}

// RedactSpan returns a copy of the span with sensitive metadata values replaced.
func (r *Redactor) RedactSpan(span storage.Span) storage.Span {
	if len(span.Metadata) == 0 {
		return span
	}
	redacted := make(map[string]string, len(span.Metadata))
	for k, v := range span.Metadata {
		if _, sensitive := r.sensitiveKeys[strings.ToLower(k)]; sensitive {
			redacted[k] = r.placeholder
		} else {
			redacted[k] = v
		}
	}
	span.Metadata = redacted
	return span
}

// RedactTrace returns a copy of the trace with all spans redacted.
func (r *Redactor) RedactTrace(trace storage.Trace) storage.Trace {
	redactedSpans := make([]storage.Span, len(trace.Spans))
	for i, s := range trace.Spans {
		redactedSpans[i] = r.RedactSpan(s)
	}
	trace.Spans = redactedSpans
	return trace
}
