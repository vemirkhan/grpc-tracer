package tracesampler

import (
	"context"
	"errors"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/user/grpc-"
	"github.cer/internal/storage"
grpc-tracer/internal/collectorHandler(_ context.Context, _) {
	return "ok", nil
}

func errHandler(_ context.Context, _ interface{}) (interface{}, error) {
	return nil, status.Error(codes.Internal, "boom")
}

func plainErrHandler(_ context.Context, _ interface{}) (interface{}, error) {
	return nil, errors.New("plain")
}

func ctxWithSpan(traceID string) context.Context {
	span := collector.Span{
		TraceID:   traceID,
		SpanID:    "span-1",
		Service:   "svc",
		Method:    "/pkg.Svc/Method",
		StartTime: time.Now(),
		Duration:  5 * time.Millisecond,
	}
	return spancontext.WithSpan(context.Background(), span)
}

func TestInterceptor_SampledSpanIsStored(t *testing.T) {
	store := storage.NewTraceStore()
	s := New(Config{SampleRate: 1.0}) // always sample

	interceptor := UnaryServerInterceptor(s, store)
	info := &grpc.UnaryServerInfo{FullMethod: "/pkg.Svc/Method"}

	_, err := interceptor(ctxWithSpan("trace-aaa"), nil, info, okHandler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	spans := store.GetTrace("trace-aaa")
	if len(spans) != 1 {
		t.Fatalf("expected 1 stored span, got %d", len(spans))
	}
}

func TestInterceptor_NotSampledSpanIsNotStored(t *testing.T) {
	store := storage.NewTraceStore()
	s := New(Config{SampleRate: 0.0}) // never sample

	interceptor := UnaryServerInterceptor(s, store)
	info := &grpc.UnaryServerInfo{FullMethod: "/pkg.Svc/Method"}

	_, _ = interceptor(ctxWithSpan("trace-bbb"), nil, info, okHandler)

	spans := store.GetTrace("trace-bbb")
	if len(spans) != 0 {
		t.Fatalf("expected 0 stored spans, got %d", len(spans))
	}
}

func TestInterceptor_HandlerErrorPropagated(t *testing.T) {
	store := storage.NewTraceStore()
	s := New(Config{SampleRate: 1.0})

	interceptor := UnaryServerInterceptor(s, store)
	info := &grpc.UnaryServerInfo{FullMethod: "/pkg.Svc/Method"}

	_, err := interceptor(ctxWithSpan("trace-ccc"), nil, info, errHandler)
	if err == nil {
		t.Fatal("expected error to be propagated")
	}
	if status.Code(err) != codes.Internal {
		t.Fatalf("expected Internal, got %v", status.Code(err))
	}
}

func TestInterceptor_NoSpanInContext_DoesNotPanic(t *testing.T) {
	store := storage.NewTraceStore()
	s := New(Config{SampleRate: 1.0})

	interceptor := UnaryServerInterceptor(s, store)
	info := &grpc.UnaryServerInfo{FullMethod: "/pkg.Svc/Method"}

	// context has no span attached
	_, err := interceptor(context.Background(), nil, info, okHandler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
