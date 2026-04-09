package interceptor_test

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/user/grpc-tracer/internal/interceptor"
	"github.com/user/grpc-tracer/internal/storage"
)

func TestPropagatingInterceptor_InjectsNewTrace(t *testing.T) {
	store := storage.NewTraceStore()
	intc := interceptor.PropagatingUnaryServerInterceptor(store)

	ctx := context.Background()
	info := &grpc.UnaryServerInfo{FullMethod: "/svc/Method"}

	var capturedCtx context.Context
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		capturedCtx = ctx
		return "ok", nil
	}

	resp, err := intc(ctx, nil, info, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "ok" {
		t.Errorf("expected 'ok', got %v", resp)
	}
	if capturedCtx == nil {
		t.Fatal("handler was not called")
	}
}

func TestPropagatingInterceptor_ExtractsExistingTrace(t *testing.T) {
	store := storage.NewTraceStore()
	intc := interceptor.PropagatingUnaryServerInterceptor(store)

	md := metadata.Pairs(
		"x-trace-id", "trace-abc",
		"x-span-id", "span-xyz",
		"x-service-name", "upstream-svc",
	)
	ctx := metadata.NewIncomingContext(context.Background(), md)
	info := &grpc.UnaryServerInfo{FullMethod: "/svc/Downstream"}

	called := false
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		called = true
		return "done", nil
	}

	_, err := intc(ctx, nil, info, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("handler was not invoked")
	}

	traces := store.GetAllTraces()
	if len(traces) == 0 {
		t.Fatal("expected at least one trace in store")
	}

	spans, ok := traces["trace-abc"]
	if !ok {
		t.Fatalf("expected trace-abc in store, got keys: %v", func() []string {
			keys := make([]string, 0, len(traces))
			for k := range traces {
				keys = append(keys, k)
			}
			return keys
		}())
	}
	if len(spans) == 0 {
		t.Fatal("expected spans under trace-abc")
	}
}
