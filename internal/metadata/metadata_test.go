package metadata_test

import (
	"context"
	"testing"

	"google.golang.org/grpc/metadata"

	tracemeta "github.com/you/grpc-tracer/internal/metadata"
)

func incomingCtx(pairs ...string) context.Context {
	md := metadata.Pairs(pairs...)
	return metadata.NewIncomingContext(context.Background(), md)
}

func TestFromIncoming_NoMetadata(t *testing.T) {
	_, ok := tracemeta.FromIncoming(context.Background())
	if ok {
		t.Fatal("expected ok=false when no metadata present")
	}
}

func TestFromIncoming_AllFields(t *testing.T) {
	ctx := incomingCtx(
		tracemeta.KeyTraceID, "tid-1",
		tracemeta.KeySpanID, "sid-1",
		tracemeta.KeyParentID, "pid-1",
		tracemeta.KeyService, "svc-a",
	)
	info, ok := tracemeta.FromIncoming(ctx)
	if !ok {
		t.Fatal("expected ok=true")
	}
	if info.TraceID != "tid-1" {
		t.Errorf("TraceID: got %q, want %q", info.TraceID, "tid-1")
	}
	if info.SpanID != "sid-1" {
		t.Errorf("SpanID: got %q, want %q", info.SpanID, "sid-1")
	}
	if info.ParentID != "pid-1" {
		t.Errorf("ParentID: got %q, want %q", info.ParentID, "pid-1")
	}
	if info.Service != "svc-a" {
		t.Errorf("Service: got %q, want %q", info.Service, "svc-a")
	}
}

func TestFromIncoming_PartialFields(t *testing.T) {
	ctx := incomingCtx(tracemeta.KeyTraceID, "tid-2")
	info, ok := tracemeta.FromIncoming(ctx)
	if !ok {
		t.Fatal("expected ok=true even with partial metadata")
	}
	if info.TraceID != "tid-2" {
		t.Errorf("TraceID: got %q, want %q", info.TraceID, "tid-2")
	}
	if info.SpanID != "" {
		t.Errorf("SpanID should be empty, got %q", info.SpanID)
	}
}

func TestToOutgoing_SetsKeys(t *testing.T) {
	info := tracemeta.TraceInfo{
		TraceID: "tid-3",
		SpanID:  "sid-3",
		Service: "svc-b",
	}
	ctx := tracemeta.ToOutgoing(context.Background(), info)
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		t.Fatal("expected outgoing metadata to be set")
	}
	assertMD := func(key, want string) {
		t.Helper()
		vals := md.Get(key)
		if len(vals) == 0 || vals[0] != want {
			t.Errorf("key %q: got %v, want %q", key, vals, want)
		}
	}
	assertMD(tracemeta.KeyTraceID, "tid-3")
	assertMD(tracemeta.KeySpanID, "sid-3")
	assertMD(tracemeta.KeyService, "svc-b")
}

func TestToOutgoing_SkipsEmptyFields(t *testing.T) {
	ctx := tracemeta.ToOutgoing(context.Background(), tracemeta.TraceInfo{})
	_, ok := metadata.FromOutgoingContext(ctx)
	if ok {
		t.Error("expected no outgoing metadata when all fields are empty")
	}
}
