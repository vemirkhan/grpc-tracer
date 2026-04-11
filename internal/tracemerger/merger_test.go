package tracemerger_test

import (
	"testing"
	"time"

	"github.com/user/grpc-tracer/internal/storage"
	"github.com/user/grpc-tracer/internal/tracemerger"
)

func makeStore(t *testing.T) *storage.TraceStore {
	t.Helper()
	return storage.NewTraceStore()
}

func addSpan(store *storage.TraceStore, traceID, spanID, service string) {
	store.AddSpan(storage.Span{
		TraceID:   traceID,
		SpanID:    spanID,
		Service:   service,
		Method:    "/svc/Method",
		StartTime: time.Now(),
		Duration:  10 * time.Millisecond,
	})
}

func TestNew_NilDestination(t *testing.T) {
	_, err := tracemerger.New(nil)
	if err == nil {
		t.Fatal("expected error for nil destination")
	}
}

func TestMerge_NoSources(t *testing.T) {
	dst := makeStore(t)
	m, _ := tracemerger.New(dst)
	_, err := m.Merge()
	if err == nil {
		t.Fatal("expected ErrNoSources")
	}
}

func TestMerge_NilSourceReturnsError(t *testing.T) {
	dst := makeStore(t)
	m, _ := tracemerger.New(dst)
	_, err := m.Merge(nil)
	if err == nil {
		t.Fatal("expected error for nil source")
	}
}

func TestMerge_CombinesSpans(t *testing.T) {
	src1 := makeStore(t)
	src2 := makeStore(t)
	addSpan(src1, "trace-1", "span-a", "serviceA")
	addSpan(src2, "trace-1", "span-b", "serviceB")

	dst := makeStore(t)
	m, _ := tracemerger.New(dst)
	n, err := m.Merge(src1, src2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 2 {
		t.Fatalf("expected 2 spans added, got %d", n)
	}
	spans := dst.GetTrace("trace-1")
	if len(spans) != 2 {
		t.Fatalf("expected 2 spans in dst, got %d", len(spans))
	}
}

func TestMerge_DeduplicatesSpanID(t *testing.T) {
	src1 := makeStore(t)
	src2 := makeStore(t)
	addSpan(src1, "trace-2", "span-dup", "serviceA")
	addSpan(src2, "trace-2", "span-dup", "serviceB") // same span ID

	dst := makeStore(t)
	m, _ := tracemerger.New(dst)
	n, err := m.Merge(src1, src2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1 span added (deduped), got %d", n)
	}
}

func TestMerge_MultipleTraces(t *testing.T) {
	src := makeStore(t)
	addSpan(src, "t1", "s1", "svcA")
	addSpan(src, "t2", "s2", "svcB")

	dst := makeStore(t)
	m, _ := tracemerger.New(dst)
	n, err := m.Merge(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 2 {
		t.Fatalf("expected 2, got %d", n)
	}
}
