package tracepartitioner_test

import (
	"context"
	"testing"
	"time"

	"github.com/example/grpc-tracer/internal/spancontext"
	"github.com/example/grpc-tracer/internal/storage"
	"github.com/example/grpc-tracer/internal/tracepartitioner"
	"google.golang.org/grpc"
)

func okHandlerI(_ context.Context, req interface{}) (interface{}, error) {
	return req, nil
}

func ctxWithSpanI(traceID, service, method string) context.Context {
	span := storage.Span{
		TraceID:   traceID,
		SpanID:    "sp1",
		Service:   service,
		Method:    method,
		StartTime: time.Now(),
		Duration:  time.Millisecond,
	}
	return spancontext.WithSpan(context.Background(), span)
}

func TestInterceptor_RoutesSpanToCorrectPartition(t *testing.T) {
	parts := make(map[string]*storage.TraceStore)
	intercept := tracepartitioner.UnaryServerInterceptor(tracepartitioner.ByService, parts)
	info := &grpc.UnaryServerInfo{FullMethod: "/svc/Method"}

	ctx := ctxWithSpanI("t1", "auth", "/Login")
	_, err := intercept(ctx, nil, info, okHandlerI)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := parts["auth"]; !ok {
		t.Fatal("expected 'auth' partition to be created")
	}
	spans := parts["auth"].GetTrace("t1")
	if len(spans) != 1 {
		t.Errorf("expected 1 span in auth partition, got %d", len(spans))
	}
}

func TestInterceptor_NoSpanInContext_DoesNotPanic(t *testing.T) {
	parts := make(map[string]*storage.TraceStore)
	intercept := tracepartitioner.UnaryServerInterceptor(tracepartitioner.ByService, parts)
	info := &grpc.UnaryServerInfo{}

	_, err := intercept(context.Background(), nil, info, okHandlerI)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(parts) != 0 {
		t.Errorf("expected no partitions, got %d", len(parts))
	}
}

func TestInterceptor_EmptyKeyFallsToDefault(t *testing.T) {
	parts := make(map[string]*storage.TraceStore)
	emptyKey := func(s storage.Span) string { return "" }
	intercept := tracepartitioner.UnaryServerInterceptor(emptyKey, parts)
	info := &grpc.UnaryServerInfo{}

	ctx := ctxWithSpanI("t1", "svc", "/M")
	_, err := intercept(ctx, nil, info, okHandlerI)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := parts["__default__"]; !ok {
		t.Fatal("expected __default__ partition")
	}
}
