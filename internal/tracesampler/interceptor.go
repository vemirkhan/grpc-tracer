// Package tracesampler provides adaptive sampling for gRPC traces.
package tracesampler

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/user/grpc-tracer/internal/spancontext"
	"github.com/user/grpc-tracer/internal/storage"
)

// UnaryServerInterceptor returns a gRPC unary server interceptor that applies
// adaptive sampling. that are not sampled are not stored. The handler is
// always called regardless of the sampling decision so that the RPC itself is
// never dropped.
func UnaryServerInterceptor(s *AdaptiveSampler, store *storage.TraceStore) grpc.UnaryServerInterceptor {
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

		// Record the outcome so the sampler can adapt its rate.
		if err != nil {
			if st, _ := status.FromError(err); st.Code() == codes.Internal {
				s.RecordError()
			} else {
				s.RecordSuccess()
			}
		} else {
			s.RecordSuccess()
		}

		if s.ShouldSample(span.TraceID) {
			store.AddSpan(span.TraceID, span)
		}

		return resp, err
	}
}
