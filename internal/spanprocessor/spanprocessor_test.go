package spanprocessor_test

import (
	"testing"

	"github.com/user/grpc-tracer/internal/collector"
	"github.com/user/grpc-tracer/internal/spanprocessor"
)

func makeSpan(service, errMsg string, tags map[string]string) collector.Span {
	return collector.Span{
		TraceID: "t1",
		SpanID:  "s1",
		Service: service,
		Error:   errMsg,
		Tags:    tags,
	}
}

func TestPipeline_Empty_PassesThrough(t *testing.T) {
	p := spanprocessor.New()
	span := makeSpan("svc", "", nil)
	out, ok := p.Process(span)
	if !ok {
		t.Fatal("expected span to pass through empty pipeline")
	}
	if out.Service != "svc" {
		t.Errorf("expected service svc, got %s", out.Service)
	}
}

func TestEnrich_AddsTag(t *testing.T) {
	p := spanprocessor.New(spanprocessor.Enrich("env", "prod"))
	span := makeSpan("svc", "", nil)
	out, ok := p.Process(span)
	if !ok {
		t.Fatal("expected span to pass")
	}
	if out.Tags["env"] != "prod" {
		t.Errorf("expected tag env=prod, got %q", out.Tags["env"])
	}
}

func TestEnrich_PreservesExistingTags(t *testing.T) {
	p := spanprocessor.New(spanprocessor.Enrich("region", "us-east"))
	span := makeSpan("svc", "", map[string]string{"user": "alice"})
	out, ok := p.Process(span)
	if !ok {
		t.Fatal("expected span to pass")
	}
	if out.Tags["user"] != "alice" {
		t.Error("existing tag lost after Enrich")
	}
	if out.Tags["region"] != "us-east" {
		t.Error("new tag not set by Enrich")
	}
}

func TestDropOnError_DropsErrorSpan(t *testing.T) {
	p := spanprocessor.New(spanprocessor.DropOnError())
	span := makeSpan("svc", "rpc error", nil)
	_, ok := p.Process(span)
	if ok {
		t.Error("expected error span to be dropped")
	}
}

func TestDropOnError_PassesCleanSpan(t *testing.T) {
	p := spanprocessor.New(spanprocessor.DropOnError())
	span := makeSpan("svc", "", nil)
	_, ok := p.Process(span)
	if !ok {
		t.Error("expected clean span to pass")
	}
}

func TestRequireService_AllowsMatchingService(t *testing.T) {
	p := spanprocessor.New(spanprocessor.RequireService("auth", "billing"))
	span := makeSpan("auth", "", nil)
	_, ok := p.Process(span)
	if !ok {
		t.Error("expected auth span to pass")
	}
}

func TestRequireService_DropsUnknownService(t *testing.T) {
	p := spanprocessor.New(spanprocessor.RequireService("auth"))
	span := makeSpan("gateway", "", nil)
	_, ok := p.Process(span)
	if ok {
		t.Error("expected gateway span to be dropped")
	}
}

func TestPipeline_StopsOnFirstDrop(t *testing.T) {
	called := false
	marker := func(s collector.Span) (collector.Span, bool) {
		called = true
		return s, true
	}
	p := spanprocessor.New(spanprocessor.DropOnError(), marker)
	span := makeSpan("svc", "boom", nil)
	p.Process(span)
	if called {
		t.Error("subsequent processor should not be called after drop")
	}
}
