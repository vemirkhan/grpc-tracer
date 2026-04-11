package tracerotator_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/example/grpc-tracer/internal/storage"
	"github.com/example/grpc-tracer/internal/tracerotator"
)

func makeSpan(traceID, spanID string) storage.Span {
	return storage.Span{
		TraceID:   traceID,
		SpanID:    spanID,
		Service:   "svc",
		Method:    "/pkg.Svc/Method",
		StartTime: time.Now(),
		Duration:  time.Millisecond,
	}
}

func TestNew_NilStoreReturnsError(t *testing.T) {
	_, err := tracerotator.New(nil, nil)
	if err == nil {
		t.Fatal("expected error for nil store")
	}
}

func TestNew_DefaultCapacity(t *testing.T) {
	store := storage.NewTraceStore()
	r, err := tracerotator.New(store, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Len() != 0 {
		t.Fatalf("expected 0, got %d", r.Len())
	}
}

func TestAddSpan_TracksDistinctTraces(t *testing.T) {
	store := storage.NewTraceStore()
	r, _ := tracerotator.New(store, &tracerotator.Options{MaxTraces: 10})

	r.AddSpan(makeSpan("trace-1", "span-1"))
	r.AddSpan(makeSpan("trace-1", "span-2"))
	r.AddSpan(makeSpan("trace-2", "span-3"))

	if got := r.Len(); got != 2 {
		t.Fatalf("expected 2 distinct traces, got %d", got)
	}
}

func TestRotator_EvictsOldestTrace(t *testing.T) {
	store := storage.NewTraceStore()
	r, _ := tracerotator.New(store, &tracerotator.Options{MaxTraces: 3})

	for i := 1; i <= 3; i++ {
		r.AddSpan(makeSpan(fmt.Sprintf("trace-%d", i), fmt.Sprintf("span-%d", i)))
	}

	if r.Len() != 3 {
		t.Fatalf("expected 3, got %d", r.Len())
	}

	// Adding a 4th trace should evict "trace-1".
	r.AddSpan(makeSpan("trace-4", "span-4"))

	if r.Len() != 3 {
		t.Fatalf("expected 3 after eviction, got %d", r.Len())
	}

	_, found := store.GetTrace("trace-1")
	if found {
		t.Error("expected trace-1 to be evicted")
	}

	_, found = store.GetTrace("trace-4")
	if !found {
		t.Error("expected trace-4 to be present")
	}
}

func TestRotator_SpansForExistingTrace_DoNotTriggerEviction(t *testing.T) {
	store := storage.NewTraceStore()
	r, _ := tracerotator.New(store, &tracerotator.Options{MaxTraces: 2})

	r.AddSpan(makeSpan("trace-1", "span-1"))
	r.AddSpan(makeSpan("trace-2", "span-2"))
	// Additional span for an already-known trace must NOT evict anything.
	r.AddSpan(makeSpan("trace-1", "span-3"))

	if r.Len() != 2 {
		t.Fatalf("expected 2, got %d", r.Len())
	}

	spans, _ := store.GetTrace("trace-1")
	if len(spans) != 2 {
		t.Fatalf("expected 2 spans in trace-1, got %d", len(spans))
	}
}
