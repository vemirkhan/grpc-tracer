package interceptor

import (
	"context"
	"time"

	"google.golang.org/grpc"

	"github.com/user/grpc-tracer/internal/collector"
	"github.com/user/grpc-tracer/internal/propagator"
	"github.com/user/grpc-tracer/internal/storage"
)

// PropagatingUnaryServerInterceptor is a gRPC unary server interceptor that
// extracts an incoming trace context (if present) and records a span into the
// collector, continuing the distributed trace chain.
func PropagatingUnaryServerInterceptor(c *collector.Collector) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		tc, ok := propagator.Extract(ctx)
		var traceID, parentSpanID string
		if ok {
			traceID = tc.TraceID
			parentSpanID = tc.SpanID
		} else {
			traceID = generateID()
		}
		spanID := generateID()

		resp, err := handler(ctx, req)

		span := storage.Span{
			TraceID:      traceID,
			SpanID:       spanID,
			ParentSpanID: parentSpanID,
			ServiceName:  info.FullMethod,
			StartTime:    start,
			Duration:     time.Since(start),
			Error:        err,
		}
		c.AddSpan(span)

		return resp, err
	}
}
