package tracecorrelator_test

import (
	"testing"
	"time"

	"github.com/example/grpc-tracer/internal/collector"
	"github.com/example/grpc-tracer/internal/storage"
	"github.com/example/grpc-tracer/internal/tracecorrelator"
)

func makeStore(t *testing.T) *storage.TraceStore {
	t.Helper()
	store := storage.NewTraceStore()
	return store
}

func addSpan(store *storage.TraceStore, traceID, spanID, service string, tags map[string]string) {
	sp := collector.Span{
		TraceID:   traceID,
		SpanID:    spanID,
		Service:   service,
		Method:    "/svc/Method",
		StartTime: time.Now(),
		Duration:  time.Millisecond,
		Tags:      tags,
	}
	store.AddSpan(sp)
}

func TestNew_NilStoreReturnsError(t *testing.T) {
	_, err := tracecorrelator.New(nil)
	if err == nil {
		t.Fatal("expected error for nil store")
	}
}

func TestNew_ValidStore(t *testing.T) {
	store := makeStore(t)
	c, err := tracecorrelator.New(store)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil correlator")
	}
}

func TestCorrelateByTag_SharedTag(t *testing.T) {
	store := makeStore(t)
	addSpan(store, "trace-1", "span-1", "svc-a", map[string]string{"user": "alice"})
	addSpan(store, "trace-2", "span-2", "svc-b", map[string]string{"user": "alice"})
	addSpan(store, "trace-3", "span-3", "svc-c", map[string]string{"user": "bob"})

	c, _ := tracecorrelator.New(store)
	results := c.CorrelateByTag("user")

	if len(results) != 1 {
		t.Fatalf("expected 1 correlation, got %d", len(results))
	}
	if results[0].Reason != "shared-tag:user" {
		t.Errorf("unexpected reason: %s", results[0].Reason)
	}
}

func TestCorrelateByTag_NoSharedTag(t *testing.T) {
	store := makeStore(t)
	addSpan(store, "trace-1", "span-1", "svc-a", map[string]string{"user": "alice"})
	addSpan(store, "trace-2", "span-2", "svc-b", map[string]string{"user": "bob"})

	c, _ := tracecorrelator.New(store)
	results := c.CorrelateByTag("user")

	if len(results) != 0 {
		t.Fatalf("expected 0 correlations, got %d", len(results))
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	store := makeStore(t)
	addSpan(store, "trace-1", "span-1", "svc-a", map[string]string{"env": "prod"})
	addSpan(store, "trace-2", "span-2", "svc-b", map[string]string{"env": "prod"})

	c, _ := tracecorrelator.New(store)
	c.CorrelateByTag("env")

	all := c.All()
	if len(all) == 0 {
		t.Fatal("expected at least one correlation")
	}
}

func TestForTrace_FiltersCorrectly(t *testing.T) {
	store := makeStore(t)
	addSpan(store, "trace-A", "span-1", "svc-a", map[string]string{"region": "us"})
	addSpan(store, "trace-B", "span-2", "svc-b", map[string]string{"region": "us"})
	addSpan(store, "trace-C", "span-3", "svc-c", map[string]string{"region": "eu"})

	c, _ := tracecorrelator.New(store)
	c.CorrelateByTag("region")

	results := c.ForTrace("trace-A")
	if len(results) != 1 {
		t.Fatalf("expected 1 result for trace-A, got %d", len(results))
	}
	results2 := c.ForTrace("trace-C")
	if len(results2) != 0 {
		t.Fatalf("expected 0 results for trace-C, got %d", len(results2))
	}
}
