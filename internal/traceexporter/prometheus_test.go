package traceexporter_test

import (
	"sync"
	"testing"
	"time"

	"github.com/grpc-tracer/internal/storage"
	"github.com/grpc-tracer/internal/traceexporter"
)

// --- fakes ---

type fakeCounter struct {
	mu     sync.Mutex
	calls  []string
}

func (f *fakeCounter) Inc(labels ...string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(labels) > 0 {
		f.calls = append(f.calls, labels[0])
	} else {
		f.calls = append(f.calls, "")
	}
}

type fakeHistogram struct {
	mu   sync.Mutex
	vals []float64
}

func (f *fakeHistogram) Observe(v float64, labels ...string) {
	f.f.mu.Lock()
	defer f.f.mu.Unlock()
	f.f.vals = append(f.f.vals, v)
}

// simpler inline fake
type simpleFakeHistogram struct {
	mu   sync.Mutex
	vals []float64
}

func (s *simpleFakeHistogram) Observe(v float64, _ ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.vals = append(s.vals, v)
}

func makeTestStore(t *testing.T) *storage.TraceStore {
	t.Helper()
	return storage.NewTraceStore()
}

func addSpan(store *storage.TraceStore, traceID, spanID, svc string, hasError bool, dur time.Duration) {
	now := time.Now()
	span := storage.Span{
		TraceID:     traceID,
		SpanID:      spanID,
		ServiceName: svc,
		StartTime:   now,
		EndTime:     now.Add(dur),
	}
	if hasError {
		span.Error = "rpc error"
	}
	store.AddSpan(traceID, span)
}

func TestFlush_EmitsSpanAndDuration(t *testing.T) {
	store := makeTestStore(t)
	addSpan(store, "t1", "s1", "svc-a", false, 10*time.Millisecond)

	ctr := &fakeCounter{}
	errCtr := &fakeCounter{}
	hist := &simpleFakeHistogram{}

	exp := traceexporter.NewPrometheusExporter(store, traceexporter.Metrics{
		SpansTotal:  ctr,
		ErrorsTotal: errCtr,
		DurationMS:  hist,
	})

	if err := exp.Flush(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ctr.calls) != 1 {
		t.Fatalf("expected 1 span count, got %d", len(ctr.calls))
	}
	if len(errCtr.calls) != 0 {
		t.Fatalf("expected 0 error counts, got %d", len(errCtr.calls))
	}
	if len(hist.vals) != 1 {
		t.Fatalf("expected 1 duration observation, got %d", len(hist.vals))
	}
}

func TestFlush_CountsErrors(t *testing.T) {
	store := makeTestStore(t)
	addSpan(store, "t2", "s1", "svc-b", true, 5*time.Millisecond)

	ctr := &fakeCounter{}
	errCtr := &fakeCounter{}
	hist := &simpleFakeHistogram{}

	exp := traceexporter.NewPrometheusExporter(store, traceexporter.Metrics{
		SpansTotal:  ctr,
		ErrorsTotal: errCtr,
		DurationMS:  hist,
	})
	exp.Flush() //nolint

	if len(errCtr.calls) != 1 {
		t.Fatalf("expected 1 error count, got %d", len(errCtr.calls))
	}
}

func TestFlush_DeduplicatesSpans(t *testing.T) {
	store := makeTestStore(t)
	addSpan(store, "t3", "s1", "svc-c", false, 2*time.Millisecond)

	ctr := &fakeCounter{}
	exp := traceexporter.NewPrometheusExporter(store, traceexporter.Metrics{
		SpansTotal:  ctr,
		ErrorsTotal: &fakeCounter{},
		DurationMS:  &simpleFakeHistogram{},
	})

	exp.Flush() //nolint
	exp.Flush() //nolint

	if len(ctr.calls) != 1 {
		t.Fatalf("expected 1 call after two flushes, got %d", len(ctr.calls))
	}
}

func TestReset_AllowsReExport(t *testing.T) {
	store := makeTestStore(t)
	addSpan(store, "t4", "s1", "svc-d", false, 1*time.Millisecond)

	ctr := &fakeCounter{}
	exp := traceexporter.NewPrometheusExporter(store, traceexporter.Metrics{
		SpansTotal:  ctr,
		ErrorsTotal: &fakeCounter{},
		DurationMS:  &simpleFakeHistogram{},
	})

	exp.Flush() //nolint
	exp.Reset()
	exp.Flush() //nolint

	if len(ctr.calls) != 2 {
		t.Fatalf("expected 2 calls after reset, got %d", len(ctr.calls))
	}
}
