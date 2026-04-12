package tracecorrelator_test

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc"

	"github.com/example/grpc-tracer/internal/collector"
	"github.com/example/grpc-tracer/internal/spancontext"
	"github.com/example/grpc-tracer/internal/storage"
	"github.com/example/grpc-tracer/internal/tracecorrelator"
)

func okHandler(_ context.Context, req interface{}) (interface{}, error) {
	return req, nil
}

func ctxWithSpan(traceID string) context.Context {
	sp := collector.Span{
		TraceID:   traceID,
		SpanID:    "span-x",
		Service:   "svc",
		StartTime: time.Now(),
		Tags:      map[string]string{"tenant": "acme"},
	}
	return spancontext.WithSpan(context.Background(), sp)
}

func TestInterceptor_RunsWithoutPanic(t *testing.T) {
	store := storage.NewTraceStore()
	store.AddSpan(collector.Span{
		TraceID: "other-trace",
		SpanID:  "s1",
		Service: "other",
		Tags:    map[string]string{"tenant": "acme"},
	})

	c, err := tracecorrelator.New(store)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	interceptor := tracecorrelator.UnaryServerInterceptor(c, "tenant")
	info := &grpc.UnaryServerInfo{FullMethod: "/svc/Method"}

	ctx := ctxWithSpan("trace-new")
	resp, err := interceptor(ctx, "req", info, okHandler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "req" {
		t.Errorf("expected response passthrough")
	}
}

func TestInterceptor_NoSpanInContext_DoesNotPanic(t *testing.T) {
	store := storage.NewTraceStore()
	c, _ := tracecorrelator.New(store)
	interceptor := tracecorrelator.UnaryServerInterceptor(c, "tenant")
	info := &grpc.UnaryServerInfo{FullMethod: "/svc/Method"}

	_, err := interceptor(context.Background(), nil, info, okHandler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInterceptor_CorrelationsRecorded(t *testing.T) {
	store := storage.NewTraceStore()
	store.AddSpan(collector.Span{
		TraceID: "existing-trace",
		SpanID:  "s1",
		Service: "svc-b",
		Tags:    map[string]string{"tenant": "beta"},
	})
	store.AddSpan(collector.Span{
		TraceID: "new-trace",
		SpanID:  "s2",
		Service: "svc-a",
		Tags:    map[string]string{"tenant": "beta"},
	})

	c, _ := tracecorrelator.New(store)
	interceptor := tracecorrelator.UnaryServerInterceptor(c, "tenant")
	info := &grpc.UnaryServerInfo{FullMethod: "/svc/Method"}

	ctx := ctxWithSpan("new-trace")
	_, _ = interceptor(ctx, nil, info, okHandler)

	all := c.All()
	if len(all) == 0 {
		t.Fatal("expected at least one correlation after interceptor ran")
	}
}
