package traceannotator

import (
	"context"

	"github.com/user/grpc-tracer/internal/spancontext"
	"google.golang.org/grpc"
)

// StaticAnnotations is a map of key-value pairs to attach to every span
// processed by UnaryServerInterceptor.
type StaticAnnotations map[string]string

// UnaryServerInterceptor returns a gRPC unary server interceptor that stamps
// a fixed set of annotations onto the current span (if one exists in the
// context) before invoking the handler.
func UnaryServerInterceptor(an *Annotator, annotations StaticAnnotations) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if sp, ok := spancontext.FromContext(ctx); ok && sp.TraceID != "" {
			for k, v := range annotations {
				// Best-effort: ignore errors (span may not be stored yet).
				_ = an.Annotate(sp.TraceID, sp.SpanID, k, v)
			}
		}
		return handler(ctx, req)
	}
}
