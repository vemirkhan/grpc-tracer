package tracepruner_test

import (
	"testing"
	"time"

	"github.com/example/grpc-tracer/internal/storage"
	"github.com/example/grpc-tracer/internal/tracepruner"
)

func makeSpan(traceID, spanID string, age time.Duration) storage.Span {
	return storage.Span{
		TraceID:   traceID,
		SpanID:    spanID,
		Service:   "svc",
		Method:    "/pkg.Svc/Method",
		StartTime: time.Now().Add(-age),
		Duration:  time.Millisecond,
	}
}

func TestPrune_RemovesOldSpans(t *testing.T) {
	store := storage.NewTraceStore()
	store.AddSpan(makeSpan("t1", "s1", 2*time.Hour))  // old
	store.AddSpan(makeSpan("t1", "s2", 10*time.Minute)) // fresh

	p := tracepruner.New(store, tracepruner.Options{MaxAge: 1 * time.Hour})
	removed := p.Prune()

	if removed != 1 {
		t.Fatalf("expected 1 removed, got %d", removed)
	}
	spans, _ := store.GetTrace("t1")
	if len(spans) != 1 || spans[0].SpanID != "s2" {
		t.Errorf("expected only fresh span to remain, got %+v", spans)
	}
}

func TestPrune_CapsSpansPerTrace(t *testing.T) {
	store := storage.NewTraceStore()
	for i := 0; i < 5; i++ {
		store.AddSpan(makeSpan("t2", string(rune('a'+i)), time.Duration(i)*time.Minute))
	}

	p := tracepruner.New(store, tracepruner.Options{
		MaxAge:           24 * time.Hour,
		MaxSpansPerTrace: 3,
	})
	removed := p.Prune()

	if removed != 2 {
		t.Fatalf("expected 2 removed, got %d", removed)
	}
	spans, _ := store.GetTrace("t2")
	if len(spans) != 3 {
		t.Errorf("expected 3 spans remaining, got %d", len(spans))
	}
}

func TestPrune_EmptyStore(t *testing.T) {
	store := storage.NewTraceStore()
	p := tracepruner.New(store, tracepruner.Options{})
	if n := p.Prune(); n != 0 {
		t.Errorf("expected 0 pruned from empty store, got %d", n)
	}
}

func TestPrune_DefaultOptions(t *testing.T) {
	store := storage.NewTraceStore()
	// Span well within default 30-minute window — should survive.
	store.AddSpan(makeSpan("t3", "s1", 5*time.Minute))

	p := tracepruner.New(store, tracepruner.Options{}) // zero → defaults
	if n := p.Prune(); n != 0 {
		t.Errorf("expected 0 pruned, got %d", n)
	}
}

func TestPrune_MultipleTraces(t *testing.T) {
	store := storage.NewTraceStore()
	store.AddSpan(makeSpan("ta", "1", 2*time.Hour))
	store.AddSpan(makeSpan("tb", "2", 2*time.Hour))
	store.AddSpan(makeSpan("tc", "3", 5*time.Minute))

	p := tracepruner.New(store, tracepruner.Options{MaxAge: 1 * time.Hour})
	if n := p.Prune(); n != 2 {
		t.Errorf("expected 2 pruned, got %d", n)
	}
}
