package traceresolver_test

import (
	"testing"
	"time"

	"github.com/user/grpc-tracer/internal/storage"
	"github.com/user/grpc-tracer/internal/traceresolver"
)

func makeStore(t *testing.T) *storage.TraceStore {
	t.Helper()
	s := storage.NewTraceStore()
	return s
}

func addSpan(t *testing.T, s *storage.TraceStore, traceID, spanID, parentID, svc string) {
	t.Helper()
	s.AddSpan(storage.Span{
		TraceID:      traceID,
		SpanID:       spanID,
		ParentSpanID: parentID,
		ServiceName:  svc,
		StartTime:    time.Now(),
		Duration:     time.Millisecond,
	})
}

func TestNew_NilStoreReturnsError(t *testing.T) {
	_, err := traceresolver.New(nil)
	if err == nil {
		t.Fatal("expected error for nil store")
	}
}

func TestNew_ValidStore(t *testing.T) {
	s := makeStore(t)
	r, err := traceresolver.New(s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r == nil {
		t.Fatal("expected non-nil resolver")
	}
}

func TestRoot_ReturnsRootSpan(t *testing.T) {
	s := makeStore(t)
	addSpan(t, s, "t1", "s1", "", "svc-a")   // root
	addSpan(t, s, "t1", "s2", "s1", "svc-b") // child

	r, _ := traceresolver.New(s)
	root, err := r.Root("t1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if root.SpanID != "s1" {
		t.Errorf("expected root span s1, got %q", root.SpanID)
	}
}

func TestRoot_TraceNotFound(t *testing.T) {
	s := makeStore(t)
	r, _ := traceresolver.New(s)
	_, err := r.Root("missing")
	if err == nil {
		t.Fatal("expected error for missing trace")
	}
}

func TestAncestors_SingleLevel(t *testing.T) {
	s := makeStore(t)
	addSpan(t, s, "t1", "s1", "", "svc-a")
	addSpan(t, s, "t1", "s2", "s1", "svc-b")

	r, _ := traceresolver.New(s)
	chain, err := r.Ancestors("t1", "s2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain) != 1 || chain[0].SpanID != "s1" {
		t.Errorf("expected [s1], got %v", chain)
	}
}

func TestAncestors_MultiLevel(t *testing.T) {
	s := makeStore(t)
	addSpan(t, s, "t1", "s1", "", "a")
	addSpan(t, s, "t1", "s2", "s1", "b")
	addSpan(t, s, "t1", "s3", "s2", "c")

	r, _ := traceresolver.New(s)
	chain, err := r.Ancestors("t1", "s3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain) != 2 {
		t.Fatalf("expected 2 ancestors, got %d", len(chain))
	}
	if chain[0].SpanID != "s1" || chain[1].SpanID != "s2" {
		t.Errorf("unexpected chain order: %v", chain)
	}
}

func TestAncestors_SpanNotFound(t *testing.T) {
	s := makeStore(t)
	addSpan(t, s, "t1", "s1", "", "a")

	r, _ := traceresolver.New(s)
	_, err := r.Ancestors("t1", "missing")
	if err == nil {
		t.Fatal("expected error for missing span")
	}
}

func TestAncestors_RootHasNoAncestors(t *testing.T) {
	s := makeStore(t)
	addSpan(t, s, "t1", "s1", "", "a")

	r, _ := traceresolver.New(s)
	chain, err := r.Ancestors("t1", "s1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain) != 0 {
		t.Errorf("expected empty ancestor chain for root, got %v", chain)
	}
}
