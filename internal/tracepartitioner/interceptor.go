package tracepartitioner

import (
	"context"

	"github.com/example/grpc-tracer/internal/spancontext"
	"github.com/example/grpc-tracer/internal/storage"
	"google.golang.org/grpc"
)

// UnaryServerInterceptor returns a gRPC unary interceptor that routes
// the completed span into the appropriate partition store determined
// by keyFn. If no span is present in the context the call passes
// through unmodified.
func UnaryServerInterceptor(
	keyFn KeyFunc,
	parts map[string]*storage.TraceStore,
) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		_ *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		resp, err := handler(ctx, req)

		span, ok := spancontext.FromContext(ctx)
		if !ok {
			return resp, err
		}

		key := keyFn(span)
		if key == "" {
			key = "__default__"
		}

		if _, exists := parts[key]; !exists {
			parts[key] = storage.NewTraceStore()
		}
		parts[key].AddSpan(span)

		return resp, err
	}
}
