package storage

import (
	"testing"
	"time"
)

func TestNewTraceStore(t *testing.T) {
	store := NewTraceStore(5 * time.Minute)
	if store == nil {
		t.Fatal("expected non-nil store")
	}
	if store.spans == nil {
		t.Fatal("expected initialized spans map")
	}
	if store.maxAge != 5*time.Minute {
		t.Errorf("expected maxAge 5m, got %v", store.maxAge)
	}
}

func TestAddSpan(t *testing.T) {
	store := NewTraceStore(5 * time.Minute)
	span := &TraceSpan{
		TraceID:   "trace-123",
		SpanID:    "span-1",
		Service:   "user-service",
		Method:    "/user.UserService/GetUser",
		StartTime: time.Now(),
		Duration:  100 * time.Millisecond,
	}

	store.AddSpan(span)

	spans := store.GetTrace("trace-123")
	if len(spans) != 1 {
		t.Errorf("expected 1 span, got %d", len(spans))
	}
	if spans[0].SpanID != "span-1" {
		t.Errorf("expected span-1, got %s", spans[0].SpanID)
	}
}

func TestGetTrace(t *testing.T) {
	store := NewTraceStore(5 * time.Minute)
	traceID := "trace-456"

	span1 := &TraceSpan{TraceID: traceID, SpanID: "span-1", Service: "api-gateway"}
	span2 := &TraceSpan{TraceID: traceID, SpanID: "span-2", Service: "user-service", ParentID: "span-1"}

	store.AddSpan(span1)
	store.AddSpan(span2)

	spans := store.GetTrace(traceID)
	if len(spans) != 2 {
		t.Errorf("expected 2 spans, got %d", len(spans))
	}
}

func TestGetAllTraces(t *testing.T) {
	store := NewTraceStore(5 * time.Minute)

	span1 := &TraceSpan{TraceID: "trace-1", SpanID: "span-1"}
	span2 := &TraceSpan{TraceID: "trace-2", SpanID: "span-2"}

	store.AddSpan(span1)
	store.AddSpan(span2)

	allTraces := store.GetAllTraces()
	if len(allTraces) != 2 {
		t.Errorf("expected 2 traces, got %d", len(allTraces))
	}
	if _, ok := allTraces["trace-1"]; !ok {
		t.Error("expected trace-1 to exist")
	}
	if _, ok := allTraces["trace-2"]; !ok {
		t.Error("expected trace-2 to exist")
	}
}

func TestGetTrace_NonExistent(t *testing.T) {
	store := NewTraceStore(5 * time.Minute)
	spans := store.GetTrace("non-existent")
	if spans != nil {
		t.Errorf("expected nil for non-existent trace, got %v", spans)
	}
}
