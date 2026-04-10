package snapshot_test

import (
	"testing"
	"time"

	"github.com/user/grpc-tracer/internal/snapshot"
	"github.com/user/grpc-tracer/internal/storage"
)

func addSpan(store *storage.TraceStore, traceID, spanID string) {
	store.AddSpan(storage.Span{
		TraceID:   traceID,
		SpanID:    spanID,
		Service:   "svc",
		Method:    "/M",
		StartTime: time.Now(),
		Duration:  time.Millisecond,
	})
}

func TestCompare_NoChanges(t *testing.T) {
	store := makeStore(t)
	addSpan(store, "t1", "s1")

	base := snapshot.Capture(store)
	next := snapshot.Capture(store)

	diff := snapshot.Compare(base, next)
	if len(diff.Added) != 0 {
		t.Errorf("expected 0 added, got %d", len(diff.Added))
	}
	if len(diff.Removed) != 0 {
		t.Errorf("expected 0 removed, got %d", len(diff.Removed))
	}
}

func TestCompare_DetectsAdded(t *testing.T) {
	store := makeStore(t)
	addSpan(store, "t1", "s1")
	base := snapshot.Capture(store)

	addSpan(store, "t2", "s2")
	next := snapshot.Capture(store)

	diff := snapshot.Compare(base, next)
	if len(diff.Added) != 1 {
		t.Errorf("expected 1 added trace, got %d", len(diff.Added))
	}
	if diff.Added[0].TraceID != "t2" {
		t.Errorf("expected added trace t2, got %s", diff.Added[0].TraceID)
	}
}

func TestCompare_DetectsRemoved(t *testing.T) {
	store1 := makeStore(t)
	addSpan(store1, "t1", "s1")
	addSpan(store1, "t2", "s2")
	base := snapshot.Capture(store1)

	store2 := makeStore(t)
	addSpan(store2, "t1", "s1")
	next := snapshot.Capture(store2)

	diff := snapshot.Compare(base, next)
	if len(diff.Removed) != 1 {
		t.Errorf("expected 1 removed trace, got %d", len(diff.Removed))
	}
	if diff.Removed[0].TraceID != "t2" {
		t.Errorf("expected removed trace t2, got %s", diff.Removed[0].TraceID)
	}
}

func TestCompare_BothEmptySnapshots(t *testing.T) {
	base := snapshot.Capture(makeStore(t))
	next := snapshot.Capture(makeStore(t))

	diff := snapshot.Compare(base, next)
	if len(diff.Added) != 0 || len(diff.Removed) != 0 {
		t.Error("expected empty diff for two empty snapshots")
	}
}
