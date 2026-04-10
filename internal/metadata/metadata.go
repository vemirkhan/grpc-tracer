// Package metadata provides helpers for reading and writing gRPC trace
// metadata from/to context values and wire-level MD maps.
package metadata

import (
	"context"

	"google.golang.org/grpc/metadata"
)

const (
	KeyTraceID  = "x-trace-id"
	KeySpanID   = "x-span-id"
	KeyParentID = "x-parent-id"
	KeyService  = "x-service-name"
)

// TraceInfo holds the identifiers carried in gRPC metadata.
type TraceInfo struct {
	TraceID  string
	SpanID   string
	ParentID string
	Service  string
}

// FromIncoming extracts TraceInfo from the incoming gRPC metadata stored in ctx.
// Returns a zero-value TraceInfo and false when no metadata is present.
func FromIncoming(ctx context.Context) (TraceInfo, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return TraceInfo{}, false
	}
	return TraceInfo{
		TraceID:  first(md, KeyTraceID),
		SpanID:   first(md, KeySpanID),
		ParentID: first(md, KeyParentID),
		Service:  first(md, KeyService),
	}, true
}

// ToOutgoing attaches TraceInfo to the outgoing gRPC metadata in ctx.
// Only non-empty fields are written.
func ToOutgoing(ctx context.Context, info TraceInfo) context.Context {
	pairs := make([]string, 0, 8)
	appendPair := func(k, v string) {
		if v != "" {
			pairs = append(pairs, k, v)
		}
	}
	appendPair(KeyTraceID, info.TraceID)
	appendPair(KeySpanID, info.SpanID)
	appendPair(KeyParentID, info.ParentID)
	appendPair(KeyService, info.Service)
	if len(pairs) == 0 {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx, pairs...)
}

// first returns the first value for key k in md, or an empty string.
func first(md metadata.MD, k string) string {
	vals := md.Get(k)
	if len(vals) == 0 {
		return ""
	}
	return vals[0]
}
