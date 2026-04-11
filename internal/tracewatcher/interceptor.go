package tracewatcher

import (
	"context"
	"time"

	"google.golang.org/grpc"

	"github.com/example/grpc-tracer/internal/spancontext"
	"github.com/example/grpc-tracer/internal/storage"
)

// UnaryServerInterceptor returns a gRPC interceptor that emits a span event
// through the Watcher after each RPC completes, if a span exists in context.
func UnaryServerInterceptor(w *Watcher, store *storage.TraceStore) grpc.UnaryServerInterceptor {
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

		span.Duration = time.Since(start)
		if err != nil {
			span.Error = err.Error()
		}

		w.WatchStore(store, span.TraceID, span)
		return resp, err
	}
}
