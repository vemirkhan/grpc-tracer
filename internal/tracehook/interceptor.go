package tracehook

import (
	"context"

	"github.com/example/grpc-tracer/internal/spancontext"
	"github.com/example/grpc-tracer/internal/storage"
	"google.golang.org/grpc"
)

// UnaryServerInterceptor returns a gRPC unary server interceptor that fires
// the registered pre- and post-hooks around every span captured from the
// incoming request context.
func UnaryServerInterceptor(h *Hook, store *storage.TraceStore) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		resp, err := handler(ctx, req)

		if span, ok := spancontext.FromContext(ctx); ok {
			_ = h.AddSpan(store, span)
		}

		return resp, err
	}
}
