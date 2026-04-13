package tracehighlighter_test

import (
	"testing"
	"time"

	"github.com/user/grpc-tracer/internal/storage"
	"github.com/user/grpc-tracer/internal/tracehighlighter"
)

func makeStore() *storage.TraceStore {
	return storage.NewTraceStore()
}

func addSpan(s *storage.TraceStore, traceID, service, method string, isErr bool) {
	s.AddSpan(storage.Span{
		TraceID:     traceID,
		SpanID:      traceID + "-" + method,
		ServiceName: service,
		Method:      method,
		Start:       time.Now(),
		Duration:    time.Millisecond,
		Error:       isErr,
	})
}

func TestNew_NilStoreReturnsError(t *testing.T) {
	_, err := tracehighlighter.New(nil, tracehighlighter.Criteria{})
	if err == nil {
		t.Fatal("expected error for nil store")
	}
}

func TestNew_ValidStore(t *testing.T) {
	h, err := tracehighlighter.New(makeStore(), tracehighlighter.Criteria{})
	if err != nil || h == nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHighlightTrace_ByService(t *testing.T) {
	s := makeStore()
	addSpan(s, "t1", "OrderService", "/Order/Get", false)
	addSpan(s, "t1", "PaymentService", "/Pay/Charge", false)

	h, _ := tracehighlighter.New(s, tracehighlighter.Criteria{ServiceName: "order"})
	n, err := h.HighlightTrace("t1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1 highlighted span, got %d", n)
	}
}

func TestHighlightTrace_OnlyErrors(t *testing.T) {
	s := makeStore()
	addSpan(s, "t2", "Svc", "/a", false)
	addSpan(s, "t2", "Svc", "/b", true)
	addSpan(s, "t2", "Svc", "/c", true)

	h, _ := tracehighlighter.New(s, tracehighlighter.Criteria{OnlyErrors: true})
	n, err := h.HighlightTrace("t2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 2 {
		t.Fatalf("expected 2 highlighted spans, got %d", n)
	}
}

func TestHighlightTrace_NonExistent(t *testing.T) {
	s := makeStore()
	h, _ := tracehighlighter.New(s, tracehighlighter.Criteria{})
	_, err := h.HighlightTrace("no-such-trace")
	if err == nil {
		t.Fatal("expected error for non-existent trace")
	}
}

func TestHighlightAll_AcrossTraces(t *testing.T) {
	s := makeStore()
	addSpan(s, "trA", "Alpha", "/rpc", false)
	addSpan(s, "trB", "Beta", "/rpc", false)
	addSpan(s, "trB", "Alpha", "/rpc2", false)

	h, _ := tracehighlighter.New(s, tracehighlighter.Criteria{ServiceName: "alpha"})
	n, err := h.HighlightAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 2 {
		t.Fatalf("expected 2 highlighted spans across all traces, got %d", n)
	}
}
