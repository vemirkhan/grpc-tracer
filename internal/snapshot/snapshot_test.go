package snapshot_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/user/grpc-tracer/internal/snapshot"
	"github.com/user/grpc-tracer/internal/storage"
)

func makeStore(t *testing.T) *storage.TraceStore {
	t.Helper()
	return storage.NewTraceStore()
}

func TestCapture_EmptyStore(t *testing.T) {
	store := makeStore(t)
	snap := snapshot.Capture(store)

	if snap == nil {
		t.Fatal("expected non-nil snapshot")
	}
	if len(snap.Traces) != 0 {
		t.Errorf("expected 0 traces, got %d", len(snap.Traces))
	}
	if snap.CapturedAt.IsZero() {
		t.Error("expected CapturedAt to be set")
	}
}

func TestCapture_WithTraces(t *testing.T) {
	store := makeStore(t)
	store.AddSpan(storage.Span{TraceID: "t1", SpanID: "s1", Service: "svc", Method: "/Test", StartTime: time.Now(), Duration: time.Millisecond})
	store.AddSpan(storage.Span{TraceID: "t2", SpanID: "s2", Service: "svc", Method: "/Other", StartTime: time.Now(), Duration: time.Millisecond})

	snap := snapshot.Capture(store)
	if len(snap.Traces) != 2 {
		t.Errorf("expected 2 traces, got %d", len(snap.Traces))
	}
}

func TestCapture_MetaTraceCount(t *testing.T) {
	store := makeStore(t)
	store.AddSpan(storage.Span{TraceID: "t1", SpanID: "s1", Service: "svc", Method: "/M", StartTime: time.Now(), Duration: time.Millisecond})

	snap := snapshot.Capture(store)
	count, ok := snap.Meta["trace_count"]
	if !ok {
		t.Fatal("expected meta key trace_count")
	}
	if count.(int) != 1 {
		t.Errorf("expected trace_count=1, got %v", count)
	}
}

func TestMarshalJSON_Valid(t *testing.T) {
	store := makeStore(t)
	snap := snapshot.Capture(store)

	b, err := json.Marshal(snap)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}
	if !strings.Contains(string(b), "captured_at") {
		t.Error("expected captured_at in JSON output")
	}
}

func TestSummary_ContainsTraceCount(t *testing.T) {
	store := makeStore(t)
	store.AddSpan(storage.Span{TraceID: "t1", SpanID: "s1", Service: "svc", Method: "/M", StartTime: time.Now(), Duration: time.Millisecond})

	snap := snapshot.Capture(store)
	summary := snap.Summary()
	if !strings.Contains(summary, "1 trace") {
		t.Errorf("expected summary to mention '1 trace', got: %s", summary)
	}
}
