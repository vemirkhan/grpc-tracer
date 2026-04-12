package tracecorrelator

import (
	"context"

	"google.golang.org/grpc"

	"github.com/example/grpc-tracer/internal/spancontext"
)

// UnaryServerInterceptor records a correlation whenever the incoming span
// carries a "correlation-tag" tag that matches another known trace.
//
// The tag key used for correlation is configurable via tagKey.
func UnaryServerInterceptor(c *Correlator, tagKey string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		resp, err := handler(ctx, req)

		// After the handler, attempt to correlate the current span.
		if sp, ok := spancontext.FromContext(ctx); ok && sp.TraceID != "" {
			// Trigger a lightweight tag-based correlation pass so that newly
			// completed spans are indexed immediately.
			c.CorrelateByTag(tagKey)
		}

		return resp, err
	}
}
