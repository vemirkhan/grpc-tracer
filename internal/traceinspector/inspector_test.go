package traceinspector_test

import (
	"testing"
	"time"

	"github.com/example/grpc-tracer/internal/storage"
	"github.com/example/grpc-tracer/internal/traceinspector"
)

func makeStore() *storage.TraceStore {
	return storage.NewTraceStore()
}

func addSpan(store *storage.TraceStore, traceID, spanID, svc, method string, dur time.Duration, errMsg string) {
	store.AddSpan(storage.Span{
		TraceID:     traceID,
		SpanID:      spanID,
		ServiceName: svc,
		Method:      method,
		Duration:    dur,
		Error:       errMsg,
	})
}

func TestNew_NilStoreReturnsError(t *testing.T) {
	_, err := traceinspector.New(nil)
	if err == nil {
		t.Fatal("expected error for nil store")
	}
}

func TestNew_ValidStore(t *testing.T) {
	insp, err := traceinspector.New(makeStore())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if insp == nil {
		t.Fatal("expected non-nil inspector")
	}
}

func TestLongestSpan_NotFound(t *testing.T) {
	insp, _ := traceinspector.New(makeStore())
	_, err := insp.LongestSpan("missing")
	if err != traceinspector.ErrTraceNotFound {
		t.Fatalf("expected ErrTraceNotFound, got %v", err)
	}
}

func TestLongestSpan_ReturnsBestDuration(t *testing.T) {
	store := makeStore()
	addSpan(store, "t1", "s1", "svcA", "/Greet", 10*time.Millisecond, "")
	addSpan(store, "t1", "s2", "svcB", "/Auth", 50*time.Millisecond, "")
	addSpan(store, "t1", "s3", "svcC", "/DB", 20*time.Millisecond, "")

	insp, _ := traceinspector.New(store)
	got, err := insp.LongestSpan("t1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.SpanID != "s2" {
		t.Errorf("expected span s2, got %s", got.SpanID)
	}
	if got.Duration != 50*time.Millisecond {
		t.Errorf("expected 50ms, got %v", got.Duration)
	}
}

func TestCriticalPath_NotFound(t *testing.T) {
	insp, _ := traceinspector.New(makeStore())
	_, err := insp.CriticalPath("none")
	if err != traceinspector.ErrTraceNotFound {
		t.Fatalf("expected ErrTraceNotFound, got %v", err)
	}
}

func TestCriticalPath_OrderedDescending(t *testing.T) {
	store := makeStore()
	addSpan(store, "t2", "a", "svcA", "/A", 5*time.Millisecond, "")
	addSpan(store, "t2", "b", "svcB", "/B", 30*time.Millisecond, "rpc error")
	addSpan(store, "t2", "c", "svcC", "/C", 15*time.Millisecond, "")

	insp, _ := traceinspector.New(store)
	path, err := insp.CriticalPath("t2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(path) != 3 {
		t.Fatalf("expected 3 summaries, got %d", len(path))
	}
	if path[0].SpanID != "b" {
		t.Errorf("expected first span b, got %s", path[0].SpanID)
	}
	if !path[0].HasError {
		t.Error("expected HasError=true for span b")
	}
	if path[2].SpanID != "a" {
		t.Errorf("expected last span a, got %s", path[2].SpanID)
	}
}
