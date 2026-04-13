package tracetopology

import (
	"sort"
	"testing"
	"time"

	"github.com/example/grpc-tracer/internal/storage"
)

func makeStore() *storage.TraceStore {
	return storage.NewTraceStore()
}

func addSpan(store *storage.TraceStore, traceID, spanID, parentID, service string) {
	store.AddSpan(storage.Span{
		TraceID:      traceID,
		SpanID:       spanID,
		ParentSpanID: parentID,
		ServiceName:  service,
		Method:       "/svc/Call",
		StartTime:    time.Now(),
		Duration:     time.Millisecond,
	})
}

func TestNew_NilStoreReturnsError(t *testing.T) {
	_, err := New(nil)
	if err == nil {
		t.Fatal("expected error for nil store")
	}
}

func TestNew_ValidStore(t *testing.T) {
	_, err := New(makeStore())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuild_EmptyStore(t *testing.T) {
	topo, _ := New(makeStore())
	g := topo.Build()
	if len(g.Edges()) != 0 {
		t.Fatalf("expected no edges, got %d", len(g.Edges()))
	}
}

func TestBuild_SingleEdge(t *testing.T) {
	store := makeStore()
	addSpan(store, "t1", "s1", "", "frontend")
	addSpan(store, "t1", "s2", "s1", "backend")

	topo, _ := New(store)
	g := topo.Build()

	edges := g.Edges()
	if len(edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(edges))
	}
	e := edges[0]
	if e.From != "frontend" || e.To != "backend" {
		t.Errorf("unexpected edge %+v", e)
	}
	if e.Calls != 1 {
		t.Errorf("expected 1 call, got %d", e.Calls)
	}
}

func TestBuild_MultipleCallsIncrementCount(t *testing.T) {
	store := makeStore()
	addSpan(store, "t1", "a", "", "svcA")
	addSpan(store, "t1", "b", "a", "svcB")
	addSpan(store, "t2", "c", "", "svcA")
	addSpan(store, "t2", "d", "c", "svcB")

	topo, _ := New(store)
	g := topo.Build()

	for _, e := range g.Edges() {
		if e.From == "svcA" && e.To == "svcB" && e.Calls != 2 {
			t.Errorf("expected 2 calls, got %d", e.Calls)
		}
	}
}

func TestServices_ReturnsAllNodes(t *testing.T) {
	store := makeStore()
	addSpan(store, "t1", "s1", "", "alpha")
	addSpan(store, "t1", "s2", "s1", "beta")
	addSpan(store, "t1", "s3", "s2", "gamma")

	topo, _ := New(store)
	g := topo.Build()

	svcs := g.Services()
	sort.Strings(svcs)
	expected := []string{"alpha", "beta", "gamma"}
	if len(svcs) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, svcs)
	}
	for i, s := range expected {
		if svcs[i] != s {
			t.Errorf("pos %d: expected %s, got %s", i, s, svcs[i])
		}
	}
}

func TestBuild_SameServiceParentChildIgnored(t *testing.T) {
	store := makeStore()
	addSpan(store, "t1", "s1", "", "svcA")
	addSpan(store, "t1", "s2", "s1", "svcA") // same service — should not create edge

	topo, _ := New(store)
	g := topo.Build()

	if len(g.Edges()) != 0 {
		t.Errorf("expected no edges for same-service spans, got %d", len(g.Edges()))
	}
}
