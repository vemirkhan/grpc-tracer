package traceannotator_test

import (
	"context"
	"testing"

	"github.com/user/grpc-tracer/internal/spancontext"
	"github.com/user/grpc-tracer/internal/storage"
	"github.com/user/grpc-tracer/internal/traceannotator"
	"google.golang.org/grpc"
)

var dummyInfo = &grpc.UnaryServerInfo{FullMethod: "/svc/Method"}

func okHandler(_ context.Context, req interface{}) (interface{}, error) {
	return req, nil
}

func TestInterceptor_AnnotatesSpanInContext(t *testing.T) {
	store := storage.NewTraceStore()
	store.AddSpan(storage.Span{TraceID: "t1", SpanID: "s1", Service: "svc", Tags: map[string]string{}})

	an := traceannotator.New(store)
	interceptor := traceannotator.UnaryServerInterceptor(an, traceannotator.StaticAnnotations{
		"deployment": "canary",
	})

	sp := storage.Span{TraceID: "t1", SpanID: "s1"}
	ctx := spancontext.WithSpan(context.Background(), sp)

	_, err := interceptor(ctx, nil, dummyInfo, okHandler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	spans := store.GetTrace("t1")
	if len(spans) == 0 {
		t.Fatal("no spans found")
	}
	if spans[0].Tags["deployment"] != "canary" {
		t.Errorf("expected deployment=canary, got %q", spans[0].Tags["deployment"])
	}
}

func TestInterceptor_NoSpanInContext_DoesNotPanic(t *testing.T) {
	store := storage.NewTraceStore()
	an := traceannotator.New(store)
	interceptor := traceannotator.UnaryServerInterceptor(an, traceannotator.StaticAnnotations{"k": "v"})

	_, err := interceptor(context.Background(), nil, dummyInfo, okHandler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInterceptor_PassesThroughHandlerResponse(t *testing.T) {
	store := storage.NewTraceStore()
	an := traceannotator.New(store)
	interceptor := traceannotator.UnaryServerInterceptor(an, traceannotator.StaticAnnotations{})

	resp, err := interceptor(context.Background(), "hello", dummyInfo, okHandler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "hello" {
		t.Errorf("expected response %q, got %v", "hello", resp)
	}
}
