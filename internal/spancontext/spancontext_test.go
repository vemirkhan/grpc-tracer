package spancontext_test

import (
	"context"
	"testing"
	"time"

	"github.com/example/grpc-tracer/internal/spancontext"
)

func TestWithSpan_AndFromContext(t *testing.T) {
	sc := spancontext.SpanContext{
		TraceID:   "trace-1",
		SpanID:    "span-1",
		ParentID:  "",
		Service:   "auth",
		Method:    "/auth.Auth/Login",
		StartTime: time.Now(),
		Tags:      map[string]string{"env": "test"},
	}

	ctx := spancontext.WithSpan(context.Background(), sc)
	got, ok := spancontext.FromContext(ctx)

	if !ok {
		t.Fatal("expected SpanContext to be found in context")
	}
	if got.TraceID != sc.TraceID {
		t.Errorf("TraceID: got %q, want %q", got.TraceID, sc.TraceID)
	}
	if got.Service != sc.Service {
		t.Errorf("Service: got %q, want %q", got.Service, sc.Service)
	}
}

func TestFromContext_Missing(t *testing.T) {
	_, ok := spancontext.FromContext(context.Background())
	if ok {
		t.Error("expected no SpanContext in empty context")
	}
}

func TestMustFromContext_ReturnsZero(t *testing.T) {
	sc := spancontext.MustFromContext(context.Background())
	if sc.TraceID != "" || sc.SpanID != "" {
		t.Errorf("expected zero-value SpanContext, got %+v", sc)
	}
}

func TestWithTag_AddsTag(t *testing.T) {
	ctx := spancontext.WithTag(context.Background(), "region", "us-east-1")
	sc, ok := spancontext.FromContext(ctx)
	if !ok {
		t.Fatal("expected SpanContext after WithTag")
	}
	if sc.Tags["region"] != "us-east-1" {
		t.Errorf("tag region: got %q, want %q", sc.Tags["region"], "us-east-1")
	}
}

func TestWithTag_PreservesExistingTags(t *testing.T) {
	sc := spancontext.SpanContext{
		TraceID: "t1",
		Tags:    map[string]string{"env": "prod"},
	}
	ctx := spancontext.WithSpan(context.Background(), sc)
	ctx = spancontext.WithTag(ctx, "version", "v2")

	got, _ := spancontext.FromContext(ctx)
	if got.Tags["env"] != "prod" {
		t.Errorf("existing tag env lost: got %q", got.Tags["env"])
	}
	if got.Tags["version"] != "v2" {
		t.Errorf("new tag version: got %q, want %q", got.Tags["version"], "v2")
	}
	if got.TraceID != "t1" {
		t.Errorf("TraceID mutated: got %q", got.TraceID)
	}
}
