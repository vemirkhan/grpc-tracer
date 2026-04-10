package tracenormalizer

import (
	"testing"
	"time"

	"github.com/user/grpc-tracer/internal/collector"
)

func makeSpan(service, method string, dur time.Duration) collector.Span {
	return collector.Span{
		TraceID:     "trace-1",
		SpanID:      "span-1",
		ServiceName: service,
		Method:      method,
		Duration:    dur,
	}
}

func TestNew_Defaults(t *testing.T) {
	n := New(Options{})
	if !n.opts.LowercaseService {
		t.Error("expected LowercaseService to default to true")
	}
	if !n.opts.TrimMethod {
		t.Error("expected TrimMethod to default to true")
	}
}

func TestNormalizeSpan_LowercaseService(t *testing.T) {
	n := New(Options{LowercaseService: true})
	s := makeSpan("OrderService", "GetOrder", time.Millisecond)
	out := n.NormalizeSpan(s)
	if out.ServiceName != "orderservice" {
		t.Errorf("expected 'orderservice', got %q", out.ServiceName)
	}
}

func TestNormalizeSpan_TrimMethod(t *testing.T) {
	n := New(Options{TrimMethod: true})
	s := makeSpan("svc", "/pkg.Service/Method", time.Millisecond)
	out := n.NormalizeSpan(s)
	if out.Method != "pkg.Service/Method" {
		t.Errorf("expected 'pkg.Service/Method', got %q", out.Method)
	}
}

func TestNormalizeSpan_DefaultDuration(t *testing.T) {
	defDur := 5 * time.Millisecond
	n := New(Options{DefaultDuration: defDur})
	s := makeSpan("svc", "Method", 0)
	out := n.NormalizeSpan(s)
	if out.Duration != defDur {
		t.Errorf("expected default duration %v, got %v", defDur, out.Duration)
	}
}

func TestNormalizeSpan_DoesNotOverrideExistingDuration(t *testing.T) {
	existing := 10 * time.Millisecond
	n := New(Options{DefaultDuration: 5 * time.Millisecond})
	s := makeSpan("svc", "Method", existing)
	out := n.NormalizeSpan(s)
	if out.Duration != existing {
		t.Errorf("expected existing duration %v, got %v", existing, out.Duration)
	}
}

func TestNormalizeSpan_OriginalUnchanged(t *testing.T) {
	n := New(Options{})
	orig := makeSpan("MyService", "/Method", 0)
	n.NormalizeSpan(orig)
	if orig.ServiceName != "MyService" {
		t.Error("original span should not be mutated")
	}
}

func TestNormalizeTrace_AllSpansNormalized(t *testing.T) {
	n := New(Options{LowercaseService: true, TrimMethod: true})
	spans := []collector.Span{
		makeSpan("Alpha", "/a.A/Call", time.Millisecond),
		makeSpan("Beta", "/b.B/Call", 2*time.Millisecond),
	}
	result := n.NormalizeTrace(spans)
	if len(result) != 2 {
		t.Fatalf("expected 2 spans, got %d", len(result))
	}
	for _, s := range result {
		if s.ServiceName != strings.ToLower(s.ServiceName) {
			t.Errorf("service name not lowercased: %q", s.ServiceName)
		}
	}
}
