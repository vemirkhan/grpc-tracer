package tracemerger

import (
	"context"

	"google.golang.org/grpc"

	"github.com/user/grpc-tracer/internal/spancontext"
	"github.com/user/grpc-tracer/internal/storage"
)

// UnaryServerInterceptor returns a gRPC server interceptor that, after the
// handler completes, merges the span recorded in the request context into the
// provided destination store using the Merger.
//
// This is useful when multiple per-request stores are collected and need to be
// consolidated into a shared store for later querying.
func UnaryServerInterceptor(m *Merger, dst *storage.TraceStore) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		_ *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		resp, err := handler(ctx, req)

		span, ok := spancontext.FromContext(ctx)
		if ok && span.SpanID != "" {
			tmp := storage.NewTraceStore()
			tmp.AddSpan(span)
			// best-effort merge; ignore merge errors at intercept time
			_, _ = m.Merge(tmp) //nolint:errcheck
		}

		return resp, err
	}
}
