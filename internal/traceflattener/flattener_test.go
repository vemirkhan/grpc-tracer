package traceflattener_test

import (
	"testing"
	"time"

	"github.com/user/grpc-tracer/internal/storage"
	"github.com/user/grpc-tracer/internal/traceflattener"
)

func makeStore(t *testing.T) *storage.TraceStore {
	t.Helper()
	ts, err := storage.NewTraceStore()
	if err != nil {
		t.Fatalf("NewTraceStore: %v", err)
	}
	return ts
}

func addSpan(t *testing.T, ts *storage.TraceStore, traceID, spanID, parentID, svc string, start time.Time, hasErr bool) {
	t.Helper()
	ts.AddSpan(storage.Span{
		TraceID:   traceID,
		SpanID:    spanID,
		ParentID:  parentID,
		Service:   svc,
		StartTime: start,
		Error:     hasErr,
	})
}

func TestNew_NilStoreReturnsError(t *testing.T) {
	_, err := traceflattener.New(nil, traceflattener.Options{})
	if err == nil {
		t.Fatal("expected error for nil store")
	}
}

func TestNew_ValidStore(t *testing.T) {
	ts := makeStore(t)
	f, err := traceflattener.New(ts, traceflattener.Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f == nil {
		t.Fatal("expected non-nil flattener")
	}
}

func TestFlatten_NonExistentTrace(t *testing.T) {
	ts := makeStore(t)
	f, _ := traceflattener.New(ts, traceflattener.Options{})
	_, err := f.Flatten("ghost")
	if err == nil {
		t.Fatal("expected error for missing trace")
	}
}

func TestFlatten_OrderedByStartTime(t *testing.T) {
	ts := makeStore(t)
	now := time.Now()
	addSpan(t, ts, "t1", "s1", "", "svcA", now, false)
	addSpan(t, ts, "t1", "s2", "s1", "svcB", now.Add(10*time.Millisecond), false)
	addSpan(t, ts, "t1", "s3", "s1", "svcC", now.Add(5*time.Millisecond), false)

	f, _ := traceflattener.New(ts, traceflattener.Options{})
	spans, err := f.Flatten("t1")
	if err != nil {
		t.Fatalf("Flatten: %v", err)
	}
	if len(spans) != 3 {
		t.Fatalf("expected 3 spans, got %d", len(spans))
	}
	// root first, then children in start-time order
	if spans[0].Span.SpanID != "s1" {
		t.Errorf("expected root s1 first, got %s", spans[0].Span.SpanID)
	}
	if spans[1].Span.SpanID != "s3" {
		t.Errorf("expected s3 second, got %s", spans[1].Span.SpanID)
	}
}

func TestFlatten_DepthAssigned(t *testing.T) {
	ts := makeStore(t)
	now := time.Now()
	addSpan(t, ts, "t2", "r", "", "svcA", now, false)
	addSpan(t, ts, "t2", "c", "r", "svcB", now.Add(time.Millisecond), false)

	f, _ := traceflattener.New(ts, traceflattener.Options{})
	spans, _ := f.Flatten("t2")
	if spans[0].Depth != 0 {
		t.Errorf("root depth want 0, got %d", spans[0].Depth)
	}
	if spans[1].Depth != 1 {
		t.Errorf("child depth want 1, got %d", spans[1].Depth)
	}
}

func TestFlatten_IncludeErrorOnly(t *testing.T) {
	ts := makeStore(t)
	now := time.Now()
	addSpan(t, ts, "t3", "ok", "", "svcA", now, false)
	addSpan(t, ts, "t3", "bad", "ok", "svcB", now.Add(time.Millisecond), true)

	f, _ := traceflattener.New(ts, traceflattener.Options{IncludeErrorOnly: true})
	spans, _ := f.Flatten("t3")
	if len(spans) != 1 {
		t.Fatalf("expected 1 error span, got %d", len(spans))
	}
	if !spans[0].Span.Error {
		t.Error("expected error span")
	}
}

func TestFlattenAll_MaxSpans(t *testing.T) {
	ts := makeStore(t)
	now := time.Now()
	for i := 0; i < 5; i++ {
		addSpan(t, ts, "tall", string(rune('a'+i)), "", "svc", now.Add(time.Duration(i)*time.Millisecond), false)
	}
	f, _ := traceflattener.New(ts, traceflattener.Options{MaxSpans: 3})
	spans := f.FlattenAll()
	if len(spans) != 3 {
		t.Errorf("expected 3 spans (MaxSpans cap), got %d", len(spans))
	}
}
