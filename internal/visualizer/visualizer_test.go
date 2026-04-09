package visualizer

import (
	"strings"
	"testing"
	"time"

	"grpc-tracer/internal/storage"
)

func TestNewVisualizer(t *testing.T) {
	store := storage.NewTraceStore()
	v := NewVisualizer(store)

	if v == nil {
		t.Fatal("expected non-nil visualizer")
	}

	if v.store != store {
		t.Error("visualizer store mismatch")
	}
}

func TestFormatTrace(t *testing.T) {
	store := storage.NewTraceStore()
	v := NewVisualizer(store)

	traceID := "trace-123"
	now := time.Now()

	// Add test spans
	span1 := &storage.Span{
		TraceID:     traceID,
		SpanID:      "span-001",
		ServiceName: "service-a",
		Method:      "/api/users",
		StartTime:   now,
		EndTime:     now.Add(100 * time.Millisecond),
		Depth:       0,
		Metadata:    map[string]string{"user": "test"},
	}

	span2 := &storage.Span{
		TraceID:     traceID,
		SpanID:      "span-002",
		ParentID:    "span-001",
		ServiceName: "service-b",
		Method:      "/api/orders",
		StartTime:   now.Add(10 * time.Millisecond),
		EndTime:     now.Add(90 * time.Millisecond),
		Depth:       1,
		Error:       "connection timeout",
	}

	store.AddSpan(span1)
	store.AddSpan(span2)

	result, err := v.FormatTrace(traceID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, traceID) {
		t.Error("result should contain trace ID")
	}

	if !strings.Contains(result, "service-a") {
		t.Error("result should contain service-a")
	}

	if !strings.Contains(result, "service-b") {
		t.Error("result should contain service-b")
	}

	if !strings.Contains(result, "connection timeout") {
		t.Error("result should contain error message")
	}
}

func TestFormatTrace_NonExistent(t *testing.T) {
	store := storage.NewTraceStore()
	v := NewVisualizer(store)

	_, err := v.FormatTrace("non-existent")
	if err == nil {
		t.Error("expected error for non-existent trace")
	}
}

func TestFormatAllTraces(t *testing.T) {
	store := storage.NewTraceStore()
	v := NewVisualizer(store)

	now := time.Now()

	store.AddSpan(&storage.Span{
		TraceID:     "trace-1",
		SpanID:      "span-1",
		ServiceName: "service-a",
		Method:      "/test",
		StartTime:   now,
		EndTime:     now.Add(50 * time.Millisecond),
	})

	store.AddSpan(&storage.Span{
		TraceID:     "trace-2",
		SpanID:      "span-2",
		ServiceName: "service-b",
		Method:      "/test2",
		StartTime:   now,
		EndTime:     now.Add(100 * time.Millisecond),
	})

	result := v.FormatAllTraces()

	if !strings.Contains(result, "Total traces: 2") {
		t.Error("result should show 2 traces")
	}

	if !strings.Contains(result, "trace-1") {
		t.Error("result should contain trace-1")
	}

	if !strings.Contains(result, "trace-2") {
		t.Error("result should contain trace-2")
	}
}
