// Package spancontext provides utilities for attaching and retrieving
// span context values from a standard context.Context, enabling
// trace metadata to flow through the call stack without explicit passing.
package spancontext

import (
	"context"
	"time"
)

// key is an unexported type for context keys in this package.
type key int

const spanKey key = iota

// SpanContext holds the metadata associated with a single tracing span.
type SpanContext struct {
	TraceID   string
	SpanID    string
	ParentID  string
	Service   string
	Method    string
	StartTime time.Time
	Tags      map[string]string
}

// WithSpan returns a new context carrying the given SpanContext.
func WithSpan(ctx context.Context, sc SpanContext) context.Context {
	return context.WithValue(ctx, spanKey, sc)
}

// FromContext retrieves the SpanContext from ctx.
// The second return value reports whether a SpanContext was found.
func FromContext(ctx context.Context) (SpanContext, bool) {
	sc, ok := ctx.Value(spanKey).(SpanContext)
	return sc, ok
}

// MustFromContext retrieves the SpanContext from ctx or returns a zero-value
// SpanContext if none is present. Callers that require a span should prefer
// FromContext and handle the missing case explicitly.
func MustFromContext(ctx context.Context) SpanContext {
	sc, _ := FromContext(ctx)
	return sc
}

// WithTag returns a new context whose SpanContext has the given tag added.
// If no SpanContext exists in ctx, a new one is created.
func WithTag(ctx context.Context, key, value string) context.Context {
	sc, _ := FromContext(ctx)
	if sc.Tags == nil {
		sc.Tags = make(map[string]string)
	}
	sc.Tags[key] = value
	return WithSpan(ctx, sc)
}
