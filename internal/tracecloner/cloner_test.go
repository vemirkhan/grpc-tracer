package tracecloner_test

import (
	"testing"
	"time"

	"github.com/user/grpc-tracer/internal/storage"
	"github.com/user/grpc-tracer/internal/tracecloner"
)

func makeStore() *storage.TraceStore {
	return storage.NewTraceStore()
}

func addSpan(store *storage.TraceStore, traceID, spanID, service string, tags map[string]string) {
	store.AddSpan(storage.Span{
		TraceID:   traceID,
		SpanID:    spanID,
		Service:   service,
		Method:    "/svc/Method",
		StartTime: time.Now(),
		Duration:  10 * time.Millisecond,
		Tags:      tags,
	})
}

func TestNew_NilSourceReturnsError(t *testing.T) {
	_, err := tracecloner.New(nil, makeStore())
	if err == nil {
		t.Fatal("expected error for nil source")
	}
}

func TestNew_NilDestinationReturnsError(t *testing.T) {
	_, err := tracecloner.New(makeStore(), nil)
	if err == nil {
		t.Fatal("expected error for nil destination")
	}
}

func TestNew_ValidStores(t *testing.T) {
	_, err := tracecloner.New(makeStore(), makeStore())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCloneTrace_CopiesSpans(t *testing.T) {
	src, dest := makeStore(), makeStore()
	addSpan(src, "trace-1", "span-1", "svc-a", map[string]string{"k": "v"})
	addSpan(src, "trace-1", "span-2", "svc-b", nil)

	c, _ := tracecloner.New(src, dest)
	if err := c.CloneTrace("trace-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	spans, ok := dest.GetTrace("trace-1")
	if !ok {
		t.Fatal("expected trace in destination")
	}
	if len(spans) != 2 {
		t.Fatalf("expected 2 spans, got %d", len(spans))
	}
}

func TestCloneTrace_NonExistent(t *testing.T) {
	c, _ := tracecloner.New(makeStore(), makeStore())
	if err := c.CloneTrace("missing"); err == nil {
		t.Fatal("expected error for missing trace")
	}
}

func TestCloneTrace_TagsAreIndependent(t *testing.T) {
	src, dest := makeStore(), makeStore()
	addSpan(src, "t1", "s1", "svc", map[string]string{"env": "prod"})

	c, _ := tracecloner.New(src, dest)
	_ = c.CloneTrace("t1")

	// Mutate the original source span tags (indirectly via a new add).
	addSpan(src, "t1", "s1", "svc", map[string]string{"env": "staging"})

	spans, _ := dest.GetTrace("t1")
	if spans[0].Tags["env"] != "prod" {
		t.Errorf("cloned tags should be independent; got %q", spans[0].Tags["env"])
	}
}

func TestCloneAll_ReturnsCount(t *testing.T) {
	src, dest := makeStore(), makeStore()
	addSpan(src, "t1", "s1", "a", nil)
	addSpan(src, "t2", "s2", "b", nil)

	c, _ := tracecloner.New(src, dest)
	count := c.CloneAll()
	if count != 2 {
		t.Fatalf("expected 2 traces cloned, got %d", count)
	}

	if _, ok := dest.GetTrace("t1"); !ok {
		t.Error("expected t1 in destination")
	}
	if _, ok := dest.GetTrace("t2"); !ok {
		t.Error("expected t2 in destination")
	}
}
