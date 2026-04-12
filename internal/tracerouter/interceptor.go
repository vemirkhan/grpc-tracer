package tracerouter

import (
	"context"
	"time"

	"google.golang.org/grpc"

	"github.com/example/grpc-tracer/internal/spancontext"
	"github.com/example/grpc-tracer/internal/storage"
)

// UnaryServerInterceptor returns a gRPC server interceptor that routes the
// completed span (retrieved from the request context) through the Router.
// It relies on spancontext.FromContext to obtain the active span; if no span
// is present the interceptor is a no-op.
func UnaryServerInterceptor(r *Router) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)

		span, ok := spancontext.FromContext(ctx)
		if !ok {
			return resp, err
		}

		// Populate timing if not already set by an upstream interceptor.
		if span.Duration == 0 {
			span.Duration = time.Since(start)
		}
		if span.StartTime.IsZero() {
			span.StartTime = start
		}
		if err != nil {
			span.Error = err.Error()
		}

		r.Route(storage.Span{
			TraceID:   span.TraceID,
			SpanID:    span.SpanID,
			Service:   span.Service,
			Method:    info.FullMethod,
			StartTime: span.StartTime,
			Duration:  span.Duration,
			Error:     span.Error,
			Tags:      span.Tags,
		})

		return resp, err
	}
}
