package propagator_test

import (
	"context"
	"testing"

	"google.golang.org/grpc/metadata"

	"github.com/user/grpc-tracer/internal/propagator"
)

func TestInject_SetsAllKeys(t *testing.T) {
	ctx := context.Background()
	tc := propagator.TraceContext{
		TraceID:      "trace-1",
		SpanID:       "span-1",
		ParentSpanID: "parent-1",
	}
	ctx = propagator.Inject(ctx, tc)
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		t.Fatal("expected outgoing metadata to be set")
	}
	assertMD(t, md, propagator.TraceIDKey, "trace-1")
	assertMD(t, md, propagator.SpanIDKey, "span-1")
	assertMD(t, md, propagator.ParentSpanIDKey, "parent-1")
}

func TestInject_SkipsEmptyFields(t *testing.T) {
	ctx := context.Background()
	tc := propagator.TraceContext{TraceID: "trace-2"}
	ctx = propagator.Inject(ctx, tc)
	md, _ := metadata.FromOutgoingContext(ctx)
	if vals := md.Get(propagator.SpanIDKey); len(vals) != 0 {
		t.Errorf("expected no span-id, got %v", vals)
	}
}

func TestExtract_ReadsFromIncoming(t *testing.T) {
	md := metadata.Pairs(
		propagator.TraceIDKey, "trace-3",
		propagator.SpanIDKey, "span-3",
		propagator.ParentSpanIDKey, "parent-3",
	)
	ctx := metadata.NewIncomingContext(context.Background(), md)
	tc, ok := propagator.Extract(ctx)
	if !ok {
		t.Fatal("expected extraction to succeed")
	}
	if tc.TraceID != "trace-3" {
		t.Errorf("TraceID: got %q, want %q", tc.TraceID, "trace-3")
	}
	if tc.SpanID != "span-3" {
		t.Errorf("SpanID: got %q, want %q", tc.SpanID, "span-3")
	}
	if tc.ParentSpanID != "parent-3" {
		t.Errorf("ParentSpanID: got %q, want %q", tc.ParentSpanID, "parent-3")
	}
}

func TestExtract_NoMetadata(t *testing.T) {
	_, ok := propagator.Extract(context.Background())
	if ok {
		t.Error("expected extraction to fail with no metadata")
	}
}

func TestExtract_MissingTraceID(t *testing.T) {
	md := metadata.Pairs(propagator.SpanIDKey, "span-x")
	ctx := metadata.NewIncomingContext(context.Background(), md)
	_, ok := propagator.Extract(ctx)
	if ok {
		t.Error("expected extraction to fail without trace-id")
	}
}

func assertMD(t *testing.T, md metadata.MD, key, want string) {
	t.Helper()
	vals := md.Get(key)
	if len(vals) == 0 || vals[0] != want {
		t.Errorf("metadata[%q]: got %v, want %q", key, vals, want)
	}
}
