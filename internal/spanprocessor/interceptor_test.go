package spanprocessor_test

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc"

	"github.com/user/grpc-tracer/internal/collector"
	"github.com/user/grpc-tracer/internal/spancontext"
	"github.com/user/grpc-tracer/internal/spanprocessor"
	"github.com/user/grpc-tracer/internal/storage"
)

func ctxWithSpan(traceID, spanID, service string) context.Context {
	span := collector.Span{
		TraceID: traceID,
		SpanID:  spanID,
		Service: service,
	}
	return spancontext.WithSpan(context.Background(), span)
}

func okHandler(_ context.Context, req interface{}) (interface{}, error) {
	return req, nil
}

func failHandler(_ context.Context, _ interface{}) (interface{}, error) {
	return nil, errors.New("rpc failed")
}

func TestInterceptor_StoresProcessedSpan(t *testing.T) {
	store := storage.NewTraceStore()
	p := spanprocessor.New(spanprocessor.Enrich("via", "interceptor"))
	interceptor := spanprocessor.UnaryServerInterceptor(p, store)

	ctx := ctxWithSpan("trace-1", "span-1", "svc")
	info := &grpc.UnaryServerInfo{FullMethod: "/svc/Method"}

	_, err := interceptor(ctx, nil, info, okHandler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	trace, ok := store.GetTrace("trace-1")
	if !ok || len(trace) == 0 {
		t.Fatal("expected span to be stored")
	}
	if trace[0].Tags["via"] != "interceptor" {
		t.Errorf("expected enriched tag, got %q", trace[0].Tags["via"])
	}
}

func TestInterceptor_DroppedSpanNotStored(t *testing.T) {
	store := storage.NewTraceStore()
	p := spanprocessor.New(spanprocessor.DropOnError())
	interceptor := spanprocessor.UnaryServerInterceptor(p, store)

	ctx := ctxWithSpan("trace-2", "span-2", "svc")
	info := &grpc.UnaryServerInfo{FullMethod: "/svc/Fail"}

	_, _ = interceptor(ctx, nil, info, failHandler)

	_, ok := store.GetTrace("trace-2")
	if ok {
		t.Error("expected dropped span to not appear in store")
	}
}

func TestInterceptor_NoSpanInContext_IsNoop(t *testing.T) {
	store := storage.NewTraceStore()
	p := spanprocessor.New()
	interceptor := spanprocessor.UnaryServerInterceptor(p, store)

	ctx := context.Background()
	info := &grpc.UnaryServerInfo{FullMethod: "/svc/Method"}

	_, err := interceptor(ctx, "req", info, okHandler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(store.GetAllTraces()) != 0 {
		t.Error("expected no traces stored when context has no span")
	}
}
