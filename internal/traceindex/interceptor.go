package traceindex

import (
	"context"

	"github.com/example/grpc-tracer/internal/spancontext"
	"github.com/example/grpc-tracer/internal/storage"
	"google.golang.org/grpc"
)

// UnaryServerInterceptor returns a gRPC server interceptor that indexes the
// span stored in the request context into the provided Index after the handler
// completes. The Index is rebuilt from the full store so that all spans are
// always reflected.
func UnaryServerInterceptor(store *storage.TraceStore, idx *Index) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		resp, err := handler(ctx, req)

		// After the handler runs, re-index if a span is present in context.
		if span, ok := spancontext.FromContext(ctx); ok && span.TraceID != "" {
			idx.Build(store)
		}

		return resp, err
	}
}
