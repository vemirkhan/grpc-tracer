package tracesummarizer_test

import (
	"testing"
	"time"

	"github.com/user/grpc-tracer/internal/storage"
	"github.com/user/grpc-tracer/internal/tracesummarizer"
)

func makeStore(t *testing.T) *storage.TraceStore {
	t.Helper()
	store := storage.NewTraceStore()
	return store
}

func addSpan(t *testing.T, store *storage.TraceStore, traceID, spanID, parentID, svc, method, errMsg string, dur time.Duration) {
	t.Helper()
	store.AddSpan(storage.Span{
		TraceID:      traceID,
		SpanID:       spanID,
		ParentSpanID: parentID,
		ServiceName:  svc,
		Method:       method,
		Error:        errMsg,
		Duration:     dur,
	})
}

func TestNew_NilStoreReturnsError(t *testing.T) {
	_, err := tracesummarizer.New(nil)
	if err == nil {
		t.Fatal("expected error for nil store")
	}
}

func TestSummarize_TraceNotFound(t *testing.T) {
	store := makeStore(t)
	s, _ := tracesummarizer.New(store)
	_, err := s.Summarize("nonexistent")
	if err == nil {
		t.Fatal("expected error for missing trace")
	}
}

func TestSummarize_SpanCount(t *testing.T) {
	store := makeStore(t)
	addSpan(t, store, "t1", "s1", "", "svcA", "/Foo", "", 10*time.Millisecond)
	addSpan(t, store, "t1", "s2", "s1", "svcB", "/Bar", "", 5*time.Millisecond)
	s, _ := tracesummarizer.New(store)
	sum, err := s.Summarize("t1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sum.SpanCount != 2 {
		t.Errorf("expected 2 spans, got %d", sum.SpanCount)
	}
}

func TestSummarize_TotalDuration(t *testing.T) {
	store := makeStore(t)
	addSpan(t, store, "t2", "s1", "", "svcA", "/X", "", 20*time.Millisecond)
	addSpan(t, store, "t2", "s2", "s1", "svcA", "/Y", "", 30*time.Millisecond)
	s, _ := tracesummarizer.New(store)
	sum, _ := s.Summarize("t2")
	if sum.TotalDuration != 50*time.Millisecond {
		t.Errorf("expected 50ms total, got %v", sum.TotalDuration)
	}
}

func TestSummarize_HasErrors(t *testing.T) {
	store := makeStore(t)
	addSpan(t, store, "t3", "s1", "", "svcA", "/Op", "rpc error", 5*time.Millisecond)
	s, _ := tracesummarizer.New(store)
	sum, _ := s.Summarize("t3")
	if !sum.HasErrors {
		t.Error("expected HasErrors to be true")
	}
}

func TestSummarize_RootMethod(t *testing.T) {
	store := makeStore(t)
	addSpan(t, store, "t4", "s1", "", "svcA", "/Root", "", 1*time.Millisecond)
	addSpan(t, store, "t4", "s2", "s1", "svcB", "/Child", "", 1*time.Millisecond)
	s, _ := tracesummarizer.New(store)
	sum, _ := s.Summarize("t4")
	if sum.RootMethod != "/Root" {
		t.Errorf("expected root method /Root, got %q", sum.RootMethod)
	}
}

func TestSummarizeAll_ReturnsAllTraces(t *testing.T) {
	store := makeStore(t)
	addSpan(t, store, "tA", "s1", "", "svc1", "/M", "", 1*time.Millisecond)
	addSpan(t, store, "tB", "s2", "", "svc2", "/N", "", 2*time.Millisecond)
	s, _ := tracesummarizer.New(store)
	summaries := s.SummarizeAll()
	if len(summaries) != 2 {
		t.Errorf("expected 2 summaries, got %d", len(summaries))
	}
}
