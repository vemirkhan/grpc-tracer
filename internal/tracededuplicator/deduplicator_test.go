package tracededuplicator_test

import (
	"testing"
	"time"

	"github.com/user/grpc-tracer/internal/storage"
	"github.com/user/grpc-tracer/internal/tracededuplicator"
)

func makeSpan(traceID, spanID string) storage.Span {
	return storage.Span{
		TraceID:   traceID,
		SpanID:    spanID,
		Service:   "svc",
		Method:    "/pkg.Service/Method",
		StartTime: time.Now(),
		Duration:  10 * time.Millisecond,
	}
}

func TestNew_NilDestReturnsError(t *testing.T) {
	_, err := tracededuplicator.New(nil)
	if err == nil {
		t.Fatal("expected error for nil dest, got nil")
	}
}

func TestNew_ValidDest(t *testing.T) {
	store := storage.NewTraceStore()
	d, err := tracededuplicator.New(store)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d == nil {
		t.Fatal("expected non-nil deduplicator")
	}
}

func TestAddSpan_AcceptsUniqueSpan(t *testing.T) {
	store := storage.NewTraceStore()
	d, _ := tracededuplicator.New(store)

	span := makeSpan("trace-1", "span-1")
	if !d.AddSpan(span) {
		t.Fatal("expected span to be accepted")
	}
	if d.SeenCount() != 1 {
		t.Fatalf("expected 1 seen, got %d", d.SeenCount())
	}
}

func TestAddSpan_RejectsDuplicateSpan(t *testing.T) {
	store := storage.NewTraceStore()
	d, _ := tracededuplicator.New(store)

	span := makeSpan("trace-1", "span-1")
	d.AddSpan(span)
	if d.AddSpan(span) {
		t.Fatal("expected duplicate span to be rejected")
	}
	if d.SeenCount() != 1 {
		t.Fatalf("expected 1 seen, got %d", d.SeenCount())
	}
}

func TestAddSpan_DifferentSpanIDsAccepted(t *testing.T) {
	store := storage.NewTraceStore()
	d, _ := tracededuplicator.New(store)

	d.AddSpan(makeSpan("trace-1", "span-1"))
	d.AddSpan(makeSpan("trace-1", "span-2"))

	if d.SeenCount() != 2 {
		t.Fatalf("expected 2 seen, got %d", d.SeenCount())
	}
}

func TestReset_ClearsSeen(t *testing.T) {
	store := storage.NewTraceStore()
	d, _ := tracededuplicator.New(store)

	d.AddSpan(makeSpan("trace-1", "span-1"))
	d.Reset()

	if d.SeenCount() != 0 {
		t.Fatalf("expected 0 after reset, got %d", d.SeenCount())
	}
	// After reset the same span should be accepted again.
	if !d.AddSpan(makeSpan("trace-1", "span-1")) {
		t.Fatal("expected span to be accepted after reset")
	}
}

func TestAddSpan_WritesToDestStore(t *testing.T) {
	store := storage.NewTraceStore()
	d, _ := tracededuplicator.New(store)

	d.AddSpan(makeSpan("trace-42", "span-99"))

	trace, ok := store.GetTrace("trace-42")
	if !ok {
		t.Fatal("expected trace to be present in dest store")
	}
	if len(trace) != 1 {
		t.Fatalf("expected 1 span in dest store, got %d", len(trace))
	}
}
