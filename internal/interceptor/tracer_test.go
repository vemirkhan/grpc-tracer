package interceptor

import (
	"context"
	"errors"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// MockTracer for testing
type MockTracer struct {
	RecordedTraces []*TraceInfo
}

func (m *MockTracer) RecordTrace(trace *TraceInfo) {
	m.RecordedTraces = append(m.RecordedTraces, trace)
}

func TestUnaryServerInterceptor_Success(t *testing.T) {
	mockTracer := &MockTracer{}
	interceptor := UnaryServerInterceptor("test-service", mockTracer)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		time.Sleep(10 * time.Millisecond)
		return "response", nil
	}

	ctx := context.Background()
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

	resp, err := interceptor(ctx, "request", info, handler)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if resp != "response" {
		t.Errorf("Expected 'response', got %v", resp)
	}
	if len(mockTracer.RecordedTraces) != 1 {
		t.Errorf("Expected 1 recorded trace, got %d", len(mockTracer.RecordedTraces))
	}

	trace := mockTracer.RecordedTraces[0]
	if trace.ServiceName != "test-service" {
		t.Errorf("Expected service name 'test-service', got %s", trace.ServiceName)
	}
	if trace.Method != "/test.Service/Method" {
		t.Errorf("Expected method '/test.Service/Method', got %s", trace.Method)
	}
	if trace.StatusCode != "OK" {
		t.Errorf("Expected status 'OK', got %s", trace.StatusCode)
	}
	if trace.Duration < 10*time.Millisecond {
		t.Errorf("Expected duration >= 10ms, got %v", trace.Duration)
	}
}

func TestUnaryServerInterceptor_WithTraceContext(t *testing.T) {
	mockTracer := &MockTracer{}
	interceptor := UnaryServerInterceptor("test-service", mockTracer)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	}

	md := metadata.Pairs("x-trace-id", "trace-123", "x-parent-span-id", "span-456")
	ctx := metadata.NewIncomingContext(context.Background(), md)
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

	_, _ = interceptor(ctx, "request", info, handler)

	trace := mockTracer.RecordedTraces[0]
	if trace.TraceID != "trace-123" {
		t.Errorf("Expected trace ID 'trace-123', got %s", trace.TraceID)
	}
	if trace.ParentSpanID != "span-456" {
		t.Errorf("Expected parent span ID 'span-456', got %s", trace.ParentSpanID)
	}
}

func TestUnaryServerInterceptor_WithError(t *testing.T) {
	mockTracer := &MockTracer{}
	interceptor := UnaryServerInterceptor("test-service", mockTracer)

	expectedErr := errors.New("handler error")
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, expectedErr
	}

	ctx := context.Background()
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

	_, err := interceptor(ctx, "request", info, handler)

	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}

	trace := mockTracer.RecordedTraces[0]
	if trace.Error != expectedErr {
		t.Errorf("Expected trace error %v, got %v", expectedErr, trace.Error)
	}
	if trace.StatusCode == "OK" {
		t.Errorf("Expected non-OK status code, got %s", trace.StatusCode)
	}
}
