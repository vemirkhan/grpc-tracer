package tracetagger_test

import (
	"testing"
	"time"

	"github.com/user/grpc-tracer/internal/storage"
	"github.com/user/grpc-tracer/internal/tracetagger"
)

func makeSpan(traceID, spanID, service, method string) storage.Span {
	return storage.Span{
		TraceID:   traceID,
		SpanID:    spanID,
		Service:   service,
		Method:    method,
		StartTime: time.Now(),
		Duration:  10 * time.Millisecond,
		Tags:      make(map[string]string),
	}
}

func newStore(t *testing.T) *storage.TraceStore {
	t.Helper()
	s := storage.NewTraceStore()
	return s
}

func TestNew_NilStoreReturnsError(t *testing.T) {
	_, err := tracetagger.New(nil)
	if err == nil {
		t.Fatal("expected error for nil store")
	}
}

func TestNew_ValidStore(t *testing.T) {
	s := newStore(t)
	tgr, err := tracetagger.New(s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tgr == nil {
		t.Fatal("expected non-nil tagger")
	}
}

func TestTagSpan_MatchingServicePrefix(t *testing.T) {
	s := newStore(t)
	tgr, _ := tracetagger.New(s)
	tgr.AddRule(tracetagger.Rule{
		ServicePrefix: "order",
		Tags:          map[string]string{"team": "commerce"},
	})

	span := makeSpan("t1", "s1", "order-service", "/Place")
	s.AddSpan(span)

	n, err := tgr.TagSpan(span)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1 tag applied, got %d", n)
	}
}

func TestTagSpan_NoMatchingRule(t *testing.T) {
	s := newStore(t)
	tgr, _ := tracetagger.New(s)
	tgr.AddRule(tracetagger.Rule{
		ServicePrefix: "payment",
		Tags:          map[string]string{"team": "finance"},
	})

	span := makeSpan("t2", "s2", "order-service", "/Place")
	n, err := tgr.TagSpan(span)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 0 {
		t.Fatalf("expected 0 tags applied, got %d", n)
	}
}

func TestTagSpan_MatchingMethodContains(t *testing.T) {
	s := newStore(t)
	tgr, _ := tracetagger.New(s)
	tgr.AddRule(tracetagger.Rule{
		MethodContains: "Payment",
		Tags:           map[string]string{"critical": "true"},
	})

	span := makeSpan("t3", "s3", "billing", "/ProcessPayment")
	n, err := tgr.TagSpan(span)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1 tag applied, got %d", n)
	}
}

func TestTagSpan_MultipleRulesApplied(t *testing.T) {
	s := newStore(t)
	tgr, _ := tracetagger.New(s)
	tgr.AddRule(tracetagger.Rule{
		ServicePrefix: "order",
		Tags:          map[string]string{"team": "commerce"},
	})
	tgr.AddRule(tracetagger.Rule{
		MethodContains: "Place",
		Tags:           map[string]string{"flow": "checkout"},
	})

	span := makeSpan("t4", "s4", "order-service", "/PlaceOrder")
	n, err := tgr.TagSpan(span)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 2 {
		t.Fatalf("expected 2 tags applied, got %d", n)
	}
}
