package spanprocessor

import (
	"context"

	"google.golang.org/grpc"

	"github.com/user/grpc-tracer/internal/spancontext"
	"github.com/user/grpc-tracer/internal/storage"
)

// UnaryServerInterceptor returns a gRPC unary server interceptor that runs the
// captured span through the given Pipeline before persisting it to the store.
// Spans dropped by the pipeline are not stored.
func UnaryServerInterceptor(p *Pipeline, store *storage.TraceStore) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		resp, err := handler(ctx, req)

		span, ok := spancontext.FromContext(ctx)
		if !ok {
			return resp, err
		}

		if err != nil {
			span.Error = err.Error()
		}

		processed, keep := p.Process(span)
		if keep {
			store.AddSpan(processed)
		}

		return resp, err
	}
}
