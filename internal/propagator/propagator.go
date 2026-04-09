// Package propagator handles trace context propagation across gRPC metadata.
package propagator

import (
	"context"

	"google.golang.org/grpc/metadata"
)

const (
	// TraceIDKey is the metadata key used to propagate trace IDs.
	TraceIDKey = "x-trace-id"
	// SpanIDKey is the metadata key used to propagate span IDs.
	SpanIDKey = "x-span-id"
	// ParentSpanIDKey is the metadata key used to propagate parent span IDs.
	ParentSpanIDKey = "x-parent-span-id"
)

// TraceContext holds the propagated trace identifiers.
type TraceContext struct {
	TraceID      string
	SpanID       string
	ParentSpanID string
}

// Inject writes the TraceContext into the outgoing gRPC metadata of ctx.
func Inject(ctx context.Context, tc TraceContext) context.Context {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}
	md = md.Copy()
	if tc.TraceID != "" {
		md.Set(TraceIDKey, tc.TraceID)
	}
	if tc.SpanID != "" {
		md.Set(SpanIDKey, tc.SpanID)
	}
	if tc.ParentSpanID != "" {
		md.Set(ParentSpanIDKey, tc.ParentSpanID)
	}
	return metadata.NewOutgoingContext(ctx, md)
}

// Extract reads TraceContext from the incoming gRPC metadata of ctx.
// Returns an empty TraceContext and false if no trace metadata is found.
func Extract(ctx context.Context) (TraceContext, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return TraceContext{}, false
	}
	tc := TraceContext{
		TraceID:      first(md, TraceIDKey),
		SpanID:       first(md, SpanIDKey),
		ParentSpanID: first(md, ParentSpanIDKey),
	}
	if tc.TraceID == "" {
		return TraceContext{}, false
	}
	return tc, true
}

func first(md metadata.MD, key string) string {
	vals := md.Get(key)
	if len(vals) == 0 {
		return ""
	}
	return vals[0]
}
