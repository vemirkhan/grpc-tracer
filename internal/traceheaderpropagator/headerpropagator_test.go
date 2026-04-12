package traceheaderpropagator_test

import (
	"net/http"
	"testing"

	"github.com/user/grpc-tracer/internal/traceheaderpropagator"
)

func TestInject_SetsAllFields(t *testing.T) {
	h := http.Header{}
	th := traceheaderpropagator.TraceHeaders{
		TraceID:      "trace-abc",
		SpanID:       "span-xyz",
		ParentSpanID: "parent-111",
		Sampled:      true,
	}
	traceheaderpropagator.Inject(h, th)

	if got := h.Get(traceheaderpropagator.HeaderTraceID); got != "trace-abc" {
		t.Errorf("expected trace-abc, got %s", got)
	}
	if got := h.Get(traceheaderpropagator.HeaderSpanID); got != "span-xyz" {
		t.Errorf("expected span-xyz, got %s", got)
	}
	if got := h.Get(traceheaderpropagator.HeaderParentSpan); got != "parent-111" {
		t.Errorf("expected parent-111, got %s", got)
	}
	if got := h.Get(traceheaderpropagator.HeaderSampled); got != "1" {
		t.Errorf("expected 1, got %s", got)
	}
}

func TestInject_SkipsEmptyFields(t *testing.T) {
	h := http.Header{}
	traceheaderpropagator.Inject(h, traceheaderpropagator.TraceHeaders{})

	if got := h.Get(traceheaderpropagator.HeaderTraceID); got != "" {
		t.Errorf("expected empty trace-id, got %s", got)
	}
	if got := h.Get(traceheaderpropagator.HeaderSpanID); got != "" {
		t.Errorf("expected empty span-id, got %s", got)
	}
}

func TestExtract_ReadsAllFields(t *testing.T) {
	h := http.Header{}
	h.Set(traceheaderpropagator.HeaderTraceID, "t1")
	h.Set(traceheaderpropagator.HeaderSpanID, "s1")
	h.Set(traceheaderpropagator.HeaderParentSpan, "p1")
	h.Set(traceheaderpropagator.HeaderSampled, "1")

	got := traceheaderpropagator.Extract(h)
	if got.TraceID != "t1" {
		t.Errorf("expected t1, got %s", got.TraceID)
	}
	if got.SpanID != "s1" {
		t.Errorf("expected s1, got %s", got.SpanID)
	}
	if got.ParentSpanID != "p1" {
		t.Errorf("expected p1, got %s", got.ParentSpanID)
	}
	if !got.Sampled {
		t.Error("expected Sampled=true")
	}
}

func TestExtract_EmptyHeaders(t *testing.T) {
	got := traceheaderpropagator.Extract(http.Header{})
	if got.TraceID != "" || got.SpanID != "" || got.ParentSpanID != "" {
		t.Error("expected all empty fields for empty headers")
	}
	if got.Sampled {
		t.Error("expected Sampled=false for missing header")
	}
}

func TestIsValid_WithTraceID(t *testing.T) {
	th := traceheaderpropagator.TraceHeaders{TraceID: "abc"}
	if !th.IsValid() {
		t.Error("expected IsValid=true when TraceID is set")
	}
}

func TestIsValid_WithoutTraceID(t *testing.T) {
	th := traceheaderpropagator.TraceHeaders{SpanID: "s1"}
	if th.IsValid() {
		t.Error("expected IsValid=false when TraceID is empty")
	}
}

func TestExtract_SampledZero(t *testing.T) {
	h := http.Header{}
	h.Set(traceheaderpropagator.HeaderSampled, "0")
	got := traceheaderpropagator.Extract(h)
	if got.Sampled {
		t.Error("expected Sampled=false for header value '0'")
	}
}
