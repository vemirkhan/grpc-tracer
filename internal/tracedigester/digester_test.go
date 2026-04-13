package tracedigester_test

import (
	"testing"
	"time"

	"github.com/mfreeman451/grpc-tracer/internal/storage"
	"github.com/mfreeman451/grpc-tracer/internal/tracedigester"
)

func makeStore(t *testing.T) *storage.TraceStore {
	t.Helper()
	return storage.NewTraceStore()
}

func addSpan(t *testing.T, store *storage.TraceStore, traceID, spanID, svc, method string, start time.Time) {
	t.Helper()
	store.AddSpan(storage.Span{
		TraceID:     traceID,
		SpanID:      spanID,
		ServiceName: svc,
		Method:      method,
		StartTime:   start,
	})
}

func TestNew_NilStoreReturnsError(t *testing.T) {
	_, err := tracedigester.New(nil)
	if err == nil {
		t.Fatal("expected error for nil store")
	}
}

func TestNew_ValidStore(t *testing.T) {
	d, err := tracedigester.New(makeStore(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d == nil {
		t.Fatal("expected non-nil Digester")
	}
}

func TestCompute_NonExistentTrace(t *testing.T) {
	d, _ := tracedigester.New(makeStore(t))
	_, err := d.Compute("no-such-trace")
	if err == nil {
		t.Fatal("expected error for missing trace")
	}
}

func TestCompute_StableDigest(t *testing.T) {
	store := makeStore(t)
	now := time.Now()
	addSpan(t, store, "t1", "s1", "svcA", "/Hello", now)
	addSpan(t, store, "t1", "s2", "svcB", "/World", now.Add(time.Millisecond))

	d, _ := tracedigester.New(store)
	dig1, err := d.Compute("t1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	dig2, err := d.Compute("t1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dig1 != dig2 {
		t.Errorf("expected stable digest, got %q vs %q", dig1, dig2)
	}
}

func TestCompute_SameShapeProducesSameDigest(t *testing.T) {
	store := makeStore(t)
	now := time.Now()
	// Two traces with identical service/method sequences.
	addSpan(t, store, "t1", "s1", "auth", "/Login", now)
	addSpan(t, store, "t2", "s2", "auth", "/Login", now.Add(time.Second))

	d, _ := tracedigester.New(store)
	d1, _ := d.Compute("t1")
	d2, _ := d.Compute("t2")
	if d1 != d2 {
		t.Errorf("same shape should yield same digest: %q vs %q", d1, d2)
	}
}

func TestCompute_DifferentShapeProducesDifferentDigest(t *testing.T) {
	store := makeStore(t)
	now := time.Now()
	addSpan(t, store, "t1", "s1", "svcA", "/Foo", now)
	addSpan(t, store, "t2", "s2", "svcB", "/Bar", now)

	d, _ := tracedigester.New(store)
	d1, _ := d.Compute("t1")
	d2, _ := d.Compute("t2")
	if d1 == d2 {
		t.Error("different shapes should yield different digests")
	}
}

func TestGroupByDigest_GroupsIdenticalShapes(t *testing.T) {
	store := makeStore(t)
	now := time.Now()
	addSpan(t, store, "t1", "s1", "svc", "/M", now)
	addSpan(t, store, "t2", "s2", "svc", "/M", now.Add(time.Second))
	addSpan(t, store, "t3", "s3", "other", "/N", now)

	d, _ := tracedigester.New(store)
	groups, err := d.GroupByDigest()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 2 {
		t.Errorf("expected 2 digest groups, got %d", len(groups))
	}
}
