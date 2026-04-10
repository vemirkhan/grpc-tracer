package tracereplay_test

import (
	"testing"
	"time"

	"github.com/your-org/grpc-tracer/internal/storage"
	"github.com/your-org/grpc-tracer/internal/tracereplay"
)

func addSpan(store *storage.TraceStore, traceID, spanID string, start time.Time) {
	store.AddSpan(storage.Span{
		TraceID:   traceID,
		SpanID:    spanID,
		Service:   "svc",
		Method:    "/pkg.Svc/Method",
		StartTime: start,
		Duration:  10 * time.Millisecond,
	})
}

func TestNew_NilSourceReturnsError(t *testing.T) {
	dst := storage.NewTraceStore()
	_, err := tracereplay.New(nil, dst, tracereplay.Options{})
	if err == nil {
		t.Fatal("expected error for nil source")
	}
}

func TestNew_NilDestinationReturnsError(t *testing.T) {
	src := storage.NewTraceStore()
	_, err := tracereplay.New(src, nil, tracereplay.Options{})
	if err == nil {
		t.Fatal("expected error for nil destination")
	}
}

func TestReplayTrace_CopiesSpans(t *testing.T) {
	src := storage.NewTraceStore()
	dst := storage.NewTraceStore()
	now := time.Now()
	addSpan(src, "trace-1", "span-a", now)
	addSpan(src, "trace-1", "span-b", now.Add(5*time.Millisecond))

	r, err := tracereplay.New(src, dst, tracereplay.Options{SpeedFactor: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	n, err := r.ReplayTrace("trace-1")
	if err != nil {
		t.Fatalf("replay error: %v", err)
	}
	if n != 2 {
		t.Fatalf("expected 2 spans replayed, got %d", n)
	}

	spans, err := dst.GetTrace("trace-1")
	if err != nil {
		t.Fatalf("dst.GetTrace: %v", err)
	}
	if len(spans) != 2 {
		t.Fatalf("expected 2 spans in dst, got %d", len(spans))
	}
}

func TestReplayTrace_NonExistent(t *testing.T) {
	src := storage.NewTraceStore()
	dst := storage.NewTraceStore()
	r, _ := tracereplay.New(src, dst, tracereplay.Options{})

	_, err := r.ReplayTrace("ghost")
	if err == nil {
		t.Fatal("expected error for unknown trace")
	}
}

func TestReplayTrace_MaxSpansLimit(t *testing.T) {
	src := storage.NewTraceStore()
	dst := storage.NewTraceStore()
	now := time.Now()
	for i := 0; i < 5; i++ {
		addSpan(src, "t", string(rune('a'+i)), now.Add(time.Duration(i)*time.Millisecond))
	}

	r, _ := tracereplay.New(src, dst, tracereplay.Options{SpeedFactor: 0, MaxSpans: 3})
	n, err := r.ReplayTrace("t")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 3 {
		t.Fatalf("expected 3 spans, got %d", n)
	}
}

func TestReplayAll_MultipleTracks(t *testing.T) {
	src := storage.NewTraceStore()
	dst := storage.NewTraceStore()
	now := time.Now()
	addSpan(src, "t1", "s1", now)
	addSpan(src, "t2", "s2", now)

	r, _ := tracereplay.New(src, dst, tracereplay.Options{SpeedFactor: 0})
	n, err := r.ReplayAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 2 {
		t.Fatalf("expected 2 total spans, got %d", n)
	}
}
