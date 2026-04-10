package logger

import (
	"context"
	"time"

	"github.com/user/grpc-tracer/internal/storage"
	"google.golang.org/grpc"
)

// UnaryServerInterceptor returns a gRPC server interceptor that logs each
// incoming RPC as a minimal Span via the provided Logger.
// serviceNme is embedded in every emitted span.
func UnaryServerInterceptor(l *Logger, serviceName string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		dur := time.Since(start)

		errMsg := ""
		if err != nil {
			errMsg = err.Error()
		}

		span := storage.Span{
			TraceID:     "-",
			SpanID:      "-",
			ServiceName: serviceName,
			Method:      info.FullMethod,
			Duration:    dur,
			Error:       errMsg,
			StartTime:   start,
		}
		l.LogSpan(span)
		return resp, err
	}
}
