package tracerouter_test

import (
	"testing"
	"time"

	"github.com/example/grpc-tracer/internal/storage"
	"github.com/example/grpc-tracer/internal/tracerouter"
)

func makeSpan(traceID, service string) storage.Span {
	return storage.Span{
		TraceID:   traceID,
		SpanID:    "span-1",
		Service:   service,
		Method:    "/svc/Method",
		StartTime: time.Now(),
		Duration:  10 * time.Millisecond,
	}
}

func TestNew_ReturnsRouter(t *testing.T) {
	r := tracerouter.New()
	if r == nil {
		t.Fatal("expected non-nil router")
	}
}

func TestAddRoute_NilRule_ReturnsError(t *testing.T) {
	r := tracerouter.New()
	dest := storage.NewTraceStore()
	if err := r.AddRoute(nil, dest); err == nil {
		t.Fatal("expected error for nil rule")
	}
}

func TestAddRoute_NilDest_ReturnsError(t *testing.T) {
	r := tracerouter.New()
	if err := r.AddRoute(func(storage.Span) bool { return true }, nil); err == nil {
		t.Fatal("expected error for nil destination")
	}
}

func TestRoute_MatchingRule_ForwardsSpan(t *testing.T) {
	r := tracerouter.New()
	dest := storage.NewTraceStore()

	_ = r.AddRoute(func(s storage.Span) bool { return s.Service == "auth" }, dest)

	span := makeSpan("trace-1", "auth")
	n := r.Route(span)

	if n != 1 {
		t.Fatalf("expected 1 destination written, got %d", n)
	}
	trace, ok := dest.GetTrace("trace-1")
	if !ok || len(trace) != 1 {
		t.Fatalf("expected span in destination store")
	}
}

func TestRoute_NoMatch_UsesFallback(t *testing.T) {
	r := tracerouter.New()
	fallback := storage.NewTraceStore()
	r.SetFallback(fallback)

	_ = r.AddRoute(func(s storage.Span) bool { return s.Service == "auth" }, storage.NewTraceStore())

	span := makeSpan("trace-2", "billing")
	n := r.Route(span)

	if n != 1 {
		t.Fatalf("expected fallback write, got %d", n)
	}
	trace, ok := fallback.GetTrace("trace-2")
	if !ok || len(trace) != 1 {
		t.Fatal("expected span in fallback store")
	}
}

func TestRoute_MultipleRulesMatch_ForwardsToAll(t *testing.T) {
	r := tracerouter.New()
	dest1 := storage.NewTraceStore()
	dest2 := storage.NewTraceStore()

	_ = r.AddRoute(func(s storage.Span) bool { return true }, dest1)
	_ = r.AddRoute(func(s storage.Span) bool { return s.Service == "auth" }, dest2)

	span := makeSpan("trace-3", "auth")
	n := r.Route(span)

	if n != 2 {
		t.Fatalf("expected 2 destinations, got %d", n)
	}
}

func TestRoute_NoMatchNoFallback_ReturnsZero(t *testing.T) {
	r := tracerouter.New()
	_ = r.AddRoute(func(s storage.Span) bool { return false }, storage.NewTraceStore())

	n := r.Route(makeSpan("trace-4", "unknown"))
	if n != 0 {
		t.Fatalf("expected 0, got %d", n)
	}
}
