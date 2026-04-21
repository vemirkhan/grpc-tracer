package tracearchiver_test

import (
	"testing"
	"time"

	"github.com/user/grpc-tracer/internal/storage"
	"github.com/user/grpc-tracer/internal/tracearchiver"
)

func makeSpan(traceID, spanID, service string, start time.Time, dur time.Duration, hasErr bool) storage.Span {
	return storage.Span{
		TraceID:   traceID,
		SpanID:    spanID,
		Service:   service,
		Method:    "/svc.Service/Method",
		StartTime: start,
		Duration:  dur,
		Error:     hasErr,
	}
}

func makeStore(spans ...storage.Span) *storage.TraceStore {
	ts := storage.NewTraceStore()
	for _, s := range spans {
		ts.AddSpan(s)
	}
	return ts
}

func TestNew_NilStoreReturnsError(t *testing.T) {
	_, err := tracearchiver.New(nil, t.TempDir())
	if err == nil {
		t.Fatal("expected error for nil store")
	}
}

func TestNew_EmptyDirReturnsError(t *testing.T) {
	ts := storage.NewTraceStore()
	_, err := tracearchiver.New(ts, "")
	if err == nil {
		t.Fatal("expected error for empty archive dir")
	}
}

func TestNew_ValidCreation(t *testing.T) {
	ts := storage.NewTraceStore()
	a, err := tracearchiver.New(ts, t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a == nil {
		t.Fatal("expected non-nil archiver")
	}
}

func TestArchive_NonExistentTrace(t *testing.T) {
	ts := storage.NewTraceStore()
	a, _ := tracearchiver.New(ts, t.TempDir())

	err := a.Archive("no-such-trace")
	if err == nil {
		t.Fatal("expected error for non-existent trace")
	}
}

func TestArchive_WritesFile(t *testing.T) {
	now := time.Now()
	span := makeSpan("trace-1", "span-1", "order-service", now, 10*time.Millisecond, false)
	ts := makeStore(span)
	dir := t.TempDir()

	a, err := tracearchiver.New(ts, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := a.Archive("trace-1"); err != nil {
		t.Fatalf("Archive failed: %v", err)
	}
}

func TestArchive_MultipleSpans(t *testing.T) {
	now := time.Now()
	spans := []storage.Span{
		makeSpan("trace-2", "span-a", "auth", now, 5*time.Millisecond, false),
		makeSpan("trace-2", "span-b", "auth", now.Add(5*time.Millisecond), 3*time.Millisecond, false),
		makeSpan("trace-2", "span-c", "db", now.Add(8*time.Millisecond), 2*time.Millisecond, true),
	}
	ts := makeStore(spans...)
	dir := t.TempDir()

	a, _ := tracearchiver.New(ts, dir)
	if err := a.Archive("trace-2"); err != nil {
		t.Fatalf("Archive failed: %v", err)
	}
}

func TestArchiveAll_EmptyStore(t *testing.T) {
	ts := storage.NewTraceStore()
	a, _ := tracearchiver.New(ts, t.TempDir())

	n, err := a.ArchiveAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 archived, got %d", n)
	}
}

func TestArchiveAll_ArchivesAllTraces(t *testing.T) {
	now := time.Now()
	spans := []storage.Span{
		makeSpan("t1", "s1", "svc-a", now, 1*time.Millisecond, false),
		makeSpan("t2", "s2", "svc-b", now, 2*time.Millisecond, false),
		makeSpan("t3", "s3", "svc-c", now, 3*time.Millisecond, true),
	}
	ts := makeStore(spans...)
	dir := t.TempDir()

	a, _ := tracearchiver.New(ts, dir)
	n, err := a.ArchiveAll()
	if err != nil {
		t.Fatalf("ArchiveAll failed: %v", err)
	}
	if n != 3 {
		t.Errorf("expected 3 archived, got %d", n)
	}
}
