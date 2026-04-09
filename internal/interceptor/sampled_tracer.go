package interceptor

import (
	"context"

	"github.com/user/grpc-tracer/internal/collector"
	"github.com/user/grpc-tracer/internal/sampler"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// SampledUnaryServerInterceptor is a gRPC unary server interceptor that
// applies a Sampler before recording a trace span. Traces that are not
// selected by the sampler pass through without any overhead.
func SampledUnaryServerInterceptor(c *collector.Collector, s *sampler.Sampler) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Extract or generate a trace ID.
		traceID := ""
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if vals := md.Get("x-trace-id"); len(vals) > 0 {
				traceID = vals[0]
			}
		}
		if traceID == "" {
			traceID = generateID()
		}

		// Delegate to the base interceptor only when the sampler approves.
		if !s.ShouldSample(traceID) {
			return handler(ctx, req)
		}

		// Inject the trace ID so the base interceptor reuses it.
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		} else {
			md = md.Copy()
		}
		md.Set("x-trace-id", traceID)
		ctx = metadata.NewIncomingContext(ctx, md)

		return UnaryServerInterceptor(c)(ctx, req, info, handler)
	}
}
