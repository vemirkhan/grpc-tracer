package tracetemplate_test

import (
	"testing"
	"time"

	"github.com/user/grpc-tracer/internal/storage"
	"github.com/user/grpc-tracer/internal/tracetemplate"
)

func makeSpan(traceID, spanID, service string) storage.Span {
	return storage.Span{
		TraceID: traceID,
		SpanID:  spanID,
		Service: service,
		Tags:    map[string]string{},
	}
}

func TestNew_ReturnsTemplate(t *testing.T) {
	tmpl, err := tracetemplate.New(tracetemplate.WithService("default-svc"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tmpl == nil {
		t.Fatal("expected non-nil template")
	}
}

func TestApply_SetsDefaultService(t *testing.T) {
	tmpl, _ := tracetemplate.New(tracetemplate.WithService("fallback"))
	span := makeSpan("t1", "s1", "")
	out := tmpl.Apply(span)
	if out.Service != "fallback" {
		t.Errorf("expected 'fallback', got %q", out.Service)
	}
}

func TestApply_PreservesExistingService(t *testing.T) {
	tmpl, _ := tracetemplate.New(tracetemplate.WithService("fallback"))
	span := makeSpan("t1", "s1", "existing")
	out := tmpl.Apply(span)
	if out.Service != "existing" {
		t.Errorf("expected 'existing', got %q", out.Service)
	}
}

func TestApply_SetsDefaultTag(t *testing.T) {
	tmpl, _ := tracetemplate.New(tracetemplate.WithTag("env", "prod"))
	span := makeSpan("t1", "s1", "svc")
	out := tmpl.Apply(span)
	if out.Tags["env"] != "prod" {
		t.Errorf("expected tag env=prod, got %q", out.Tags["env"])
	}
}

func TestApply_DoesNotOverwriteExistingTag(t *testing.T) {
	tmpl, _ := tracetemplate.New(tracetemplate.WithTag("env", "prod"))
	span := makeSpan("t1", "s1", "svc")
	span.Tags["env"] = "staging"
	out := tmpl.Apply(span)
	if out.Tags["env"] != "staging" {
		t.Errorf("expected tag env=staging, got %q", out.Tags["env"])
	}
}

func TestApply_MinDuration_BumpsShortSpan(t *testing.T) {
	tmpl, _ := tracetemplate.New(tracetemplate.WithMinDuration(10 * time.Millisecond))
	span := makeSpan("t1", "s1", "svc")
	span.Duration = 1 * time.Millisecond
	out := tmpl.Apply(span)
	if out.Duration != 10*time.Millisecond {
		t.Errorf("expected 10ms, got %v", out.Duration)
	}
}

func TestApply_MinDuration_DoesNotShortenLongSpan(t *testing.T) {
	tmpl, _ := tracetemplate.New(tracetemplate.WithMinDuration(10 * time.Millisecond))
	span := makeSpan("t1", "s1", "svc")
	span.Duration = 50 * time.Millisecond
	out := tmpl.Apply(span)
	if out.Duration != 50*time.Millisecond {
		t.Errorf("expected 50ms, got %v", out.Duration)
	}
}

func TestApplyToTrace_NotFound(t *testing.T) {
	store := storage.NewTraceStore()
	tmpl, _ := tracetemplate.New()
	err := tmpl.ApplyToTrace(store, "missing")
	if err == nil {
		t.Fatal("expected error for missing trace")
	}
}

func TestApplyToTrace_AppliesTemplate(t *testing.T) {
	store := storage.NewTraceStore()
	span := makeSpan("trace1", "span1", "")
	store.AddSpan(span)

	tmpl, _ := tracetemplate.New(
		tracetemplate.WithService("auto-svc"),
		tracetemplate.WithTag("region", "us-east"),
	)
	if err := tmpl.ApplyToTrace(store, "trace1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	spans, _ := store.GetTrace("trace1")
	if len(spans) == 0 {
		t.Fatal("expected spans after apply")
	}
	for _, s := range spans {
		if s.Service != "auto-svc" {
			t.Errorf("expected service 'auto-svc', got %q", s.Service)
		}
		if s.Tags["region"] != "us-east" {
			t.Errorf("expected tag region=us-east, got %q", s.Tags["region"])
		}
	}
}
