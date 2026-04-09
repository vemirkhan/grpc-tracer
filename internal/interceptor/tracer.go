package interceptor

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// TraceInfo holds information about a gRPC call
type TraceInfo struct {
	TraceID     string
	SpanID      string
	ParentSpanID string
	ServiceName string
	Method      string
	StartTime   time.Time
	EndTime     time.Time
	Duration    time.Duration
	StatusCode  string
	Error       error
}

// Tracer defines the interface for trace collection
type Tracer interface {
	RecordTrace(trace *TraceInfo)
}

// UnaryServerInterceptor creates a gRPC unary server interceptor for tracing
func UnaryServerInterceptor(serviceName string, tracer Tracer) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		traceInfo := &TraceInfo{
			ServiceName: serviceName,
			Method:      info.FullMethod,
			StartTime:   time.Now(),
		}

		// Extract trace context from metadata
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if traceIDs := md.Get("x-trace-id"); len(traceIDs) > 0 {
				traceInfo.TraceID = traceIDs[0]
			}
			if spanIDs := md.Get("x-parent-span-id"); len(spanIDs) > 0 {
				traceInfo.ParentSpanID = spanIDs[0]
			}
		}

		// Generate new span ID if not present
		if traceInfo.TraceID == "" {
			traceInfo.TraceID = generateID()
		}
		traceInfo.SpanID = generateID()

		// Call the handler
		resp, err := handler(ctx, req)

		// Record trace information
		traceInfo.EndTime = time.Now()
		traceInfo.Duration = traceInfo.EndTime.Sub(traceInfo.StartTime)
		traceInfo.Error = err
		if err != nil {
			traceInfo.StatusCode = status.Code(err).String()
		} else {
			traceInfo.StatusCode = "OK"
		}

		tracer.RecordTrace(traceInfo)

		return resp, err
	}
}

// generateID generates a simple unique ID for demonstration
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
