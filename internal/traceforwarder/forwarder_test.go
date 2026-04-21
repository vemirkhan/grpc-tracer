package traceforwarder_test

import (
	"testing"
	"time"

	"github.com/example/grpc-tracer/internal/storage"
	"github.com/example/grpc-tracer/internal/traceforwarder"
)

func makeSpan(traceID, spanID, service string, hasErr bool) storage.Span {
	return storage.Span{
		TraceID:   traceID,
		SpanID:    spanID,
		Service:   service,
		Method:    "/svc/Method",
		StartTime: time.Now(),
		Duration:  10 * time.Millisecond,
		Error:     hasErr,
	}
}

func makeStore(spans ...storage.Span) *storage.TraceStore {
	s := storage.NewTraceStore()
	for _, sp := range spans {
		s.AddSpan(sp)
	}
	return s
}

func TestNew_NilSourceReturnsError(t *testing.T) {
	_, err := traceforwarder.New(nil, nil, storage.NewTraceStore())
	if err == nil {
		t.Fatal("expected error for nil source")
	}
}

func TestNew_NilDestinationReturnsError(t *testing.T) {
	src := storage.NewTraceStore()
	_, err := traceforwarder.New(src, nil, nil)
	if err == nil {
		t.Fatal("expected error for nil destination")
	}
}

func TestForwardTrace_CopiesSpans(t *testing.T) {
	span := makeSpan("trace-1", "span-1", "svcA", false)
	src := makeStore(span)
	dst := storage.NewTraceStore()

	f, err := traceforwarder.New(src, nil, dst)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	n, err := f.ForwardTrace("trace-1")
	if err != nil {
		t.Fatalf("ForwardTrace error: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1 forwarded span, got %d", n)
	}

	spans, ok := dst.GetTrace("trace-1")
	if !ok || len(spans) != 1 {
		t.Fatalf("expected 1 span in destination, got %d", len(spans))
	}
}

func TestForwardTrace_NonExistent(t *testing.T) {
	src := storage.NewTraceStore()
	dst := storage.NewTraceStore()
	f, _ := traceforwarder.New(src, nil, dst)

	_, err := f.ForwardTrace("missing")
	if err == nil {
		t.Fatal("expected error for non-existent trace")
	}
}

func TestForwardTrace_PredicateFilters(t *testing.T) {
	spans := []storage.Span{
		makeSpan("t1", "s1", "svcA", false),
		makeSpan("t1", "s2", "svcA", true), // error span – should be filtered out
	}
	src := makeStore(spans...)
	dst := storage.NewTraceStore()

	noErrors := func(s storage.Span) bool { return !s.Error }
	f, _ := traceforwarder.New(src, noErrors, dst)

	n, err := f.ForwardTrace("t1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1 forwarded span, got %d", n)
	}
}

func TestForwardAll_ForwardsEveryTrace(t *testing.T) {
	src := makeStore(
		makeSpan("t1", "s1", "svcA", false),
		makeSpan("t2", "s2", "svcB", false),
	)
	dst := storage.NewTraceStore()
	f, _ := traceforwarder.New(src, nil, dst)

	n, err := f.ForwardAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 2 {
		t.Fatalf("expected 2 total forwarded spans, got %d", n)
	}
}

func TestAddDestination_NilReturnsError(t *testing.T) {
	src := storage.NewTraceStore()
	f, _ := traceforwarder.New(src, nil)
	if err := f.AddDestination(nil); err == nil {
		t.Fatal("expected error when adding nil destination")
	}
}
