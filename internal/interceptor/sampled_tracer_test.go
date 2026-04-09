package interceptor_test

import (
	"context"
	"testing"

	"github.com/user/grpc-tracer/internal/collector"
	"github.com/user/grpc-tracer/internal/interceptor"
	"github.com/user/grpc-tracer/internal/sampler"
	"github.com/user/grpc-tracer/internal/storage"
	"google.golang.org/grpc"
)

func noopHandler(_ context.Context, req interface{}) (interface{}, error) {
	return req, nil
}

func TestSampledInterceptor_AlwaysSample(t *testing.T) {
	store := storage.NewTraceStore()
	col := collector.NewCollector(store)
	s := sampler.New(sampler.AlwaysSample, 1.0)
	interc := interceptor.SampledUnaryServerInterceptor(col, s)

	info := &grpc.UnaryServerInfo{FullMethod: "/svc/Method"}
	_, err := interc(context.Background(), "req", info, noopHandler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	traces := store.GetAllTraces()
	if len(traces) == 0 {
		t.Error("expected at least one trace to be recorded")
	}
}

func TestSampledInterceptor_NeverSample(t *testing.T) {
	store := storage.NewTraceStore()
	col := collector.NewCollector(store)
	s := sampler.New(sampler.NeverSample, 0.0)
	interc := interceptor.SampledUnaryServerInterceptor(col, s)

	info := &grpc.UnaryServerInfo{FullMethod: "/svc/Method"}
	for i := 0; i < 5; i++ {
		_, err := interc(context.Background(), "req", info, noopHandler)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	traces := store.GetAllTraces()
	if len(traces) != 0 {
		t.Errorf("expected no traces to be recorded, got %d", len(traces))
	}
}

func TestSampledInterceptor_HandlerResponsePassedThrough(t *testing.T) {
	store := storage.NewTraceStore()
	col := collector.NewCollector(store)
	s := sampler.New(sampler.AlwaysSample, 1.0)
	interc := interceptor.SampledUnaryServerInterceptor(col, s)

	info := &grpc.UnaryServerInfo{FullMethod: "/svc/Echo"}
	resp, err := interc(context.Background(), "hello", info, noopHandler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "hello" {
		t.Errorf("expected response 'hello', got %v", resp)
	}
}
