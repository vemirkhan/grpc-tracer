package traceenricher_test

import (
	"testing"

	"github.com/example/grpc-tracer/internal/storage"
	"github.com/example/grpc-tracer/internal/traceenricher"
)

func makeSpan(service string, tags map[string]string) storage.Span {
	return storage.Span{
		TraceID:     "trace-1",
		SpanID:      "span-1",
		ServiceName: service,
		Tags:        tags,
	}
}

func TestEnrich_NoRules_PassesThrough(t *testing.T) {
	e := traceenricher.New()
	span := makeSpan("OrderService", nil)
	out := e.Enrich(span)
	if out.ServiceName != "OrderService" {
		t.Fatalf("expected OrderService, got %s", out.ServiceName)
	}
}

func TestWithStaticTag_SetsTag(t *testing.T) {
	e := traceenricher.New(traceenricher.WithStaticTag("env", "production"))
	span := makeSpan("svc", nil)
	out := e.Enrich(span)
	if out.Tags["env"] != "production" {
		t.Fatalf("expected production, got %s", out.Tags["env"])
	}
}

func TestWithStaticTag_OverwritesExisting(t *testing.T) {
	e := traceenricher.New(traceenricher.WithStaticTag("env", "staging"))
	span := makeSpan("svc", map[string]string{"env": "dev"})
	out := e.Enrich(span)
	if out.Tags["env"] != "staging" {
		t.Fatalf("expected staging, got %s", out.Tags["env"])
	}
}

func TestNormalizeService_Lowercases(t *testing.T) {
	e := traceenricher.New(traceenricher.NormalizeService())
	span := makeSpan("  OrderService  ", nil)
	out := e.Enrich(span)
	if out.ServiceName != "orderservice" {
		t.Fatalf("expected orderservice, got %q", out.ServiceName)
	}
}

func TestDefaultIfEmpty_SetsWhenAbsent(t *testing.T) {
	e := traceenricher.New(traceenricher.DefaultIfEmpty("region", "us-east-1"))
	span := makeSpan("svc", nil)
	out := e.Enrich(span)
	if out.Tags["region"] != "us-east-1" {
		t.Fatalf("expected us-east-1, got %s", out.Tags["region"])
	}
}

func TestDefaultIfEmpty_DoesNotOverwriteExisting(t *testing.T) {
	e := traceenricher.New(traceenricher.DefaultIfEmpty("region", "us-east-1"))
	span := makeSpan("svc", map[string]string{"region": "eu-west-1"})
	out := e.Enrich(span)
	if out.Tags["region"] != "eu-west-1" {
		t.Fatalf("expected eu-west-1, got %s", out.Tags["region"])
	}
}

func TestEnrich_MultipleRules_AppliedInOrder(t *testing.T) {
	e := traceenricher.New(
		traceenricher.NormalizeService(),
		traceenricher.WithStaticTag("env", "test"),
		traceenricher.DefaultIfEmpty("region", "us-west-2"),
	)
	span := makeSpan("MyService", nil)
	out := e.Enrich(span)
	if out.ServiceName != "myservice" {
		t.Errorf("service: expected myservice, got %s", out.ServiceName)
	}
	if out.Tags["env"] != "test" {
		t.Errorf("env: expected test, got %s", out.Tags["env"])
	}
	if out.Tags["region"] != "us-west-2" {
		t.Errorf("region: expected us-west-2, got %s", out.Tags["region"])
	}
}
