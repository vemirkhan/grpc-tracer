package baggage_test

import (
	"context"
	"testing"

	"google.golang.org/grpc/metadata"

	"github.com/user/grpc-tracer/internal/baggage"
)

// incomingCtx converts outgoing metadata to incoming metadata so that
// Extract/Get can read what Inject wrote (simulates the wire hop).
func incomingCtx(ctx context.Context) context.Context {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		return ctx
	}
	return metadata.NewIncomingContext(context.Background(), md)
}

func TestInject_SetsAllKeys(t *testing.T) {
	b := baggage.Baggage{"request-id": "abc123", "user-id": "u42"}
	ctx := baggage.Inject(context.Background(), b)
	ctx = incomingCtx(ctx)

	got := baggage.Extract(ctx)
	if got["request-id"] != "abc123" {
		t.Errorf("expected request-id=abc123, got %q", got["request-id"])
	}
	if got["user-id"] != "u42" {
		t.Errorf("expected user-id=u42, got %q", got["user-id"])
	}
}

func TestInject_SkipsEmptyValues(t *testing.T) {
	b := baggage.Baggage{"key": "", "valid": "yes"}
	ctx := baggage.Inject(context.Background(), b)
	ctx = incomingCtx(ctx)

	got := baggage.Extract(ctx)
	if _, ok := got["key"]; ok {
		t.Error("expected empty-value key to be skipped")
	}
	if got["valid"] != "yes" {
		t.Errorf("expected valid=yes, got %q", got["valid"])
	}
}

func TestExtract_NoMetadata(t *testing.T) {
	got := baggage.Extract(context.Background())
	if len(got) != 0 {
		t.Errorf("expected empty baggage, got %v", got)
	}
}

func TestExtract_IgnoresNonBaggageKeys(t *testing.T) {
	md := metadata.Pairs("x-trace-id", "tid1", "baggage-env", "prod")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	got := baggage.Extract(ctx)
	if _, ok := got["x-trace-id"]; ok {
		t.Error("non-baggage key should not be extracted")
	}
	if got["env"] != "prod" {
		t.Errorf("expected env=prod, got %q", got["env"])
	}
}

func TestGet_ReturnsValue(t *testing.T) {
	md := metadata.Pairs("baggage-region", "us-east-1")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	if v := baggage.Get(ctx, "region"); v != "us-east-1" {
		t.Errorf("expected us-east-1, got %q", v)
	}
}

func TestGet_MissingKey(t *testing.T) {
	if v := baggage.Get(context.Background(), "missing"); v != "" {
		t.Errorf("expected empty string, got %q", v)
	}
}

func TestInject_EmptyBaggage(t *testing.T) {
	original := context.Background()
	result := baggage.Inject(original, baggage.Baggage{})
	// Should return the same context without modification.
	if _, ok := metadata.FromOutgoingContext(result); ok {
		t.Error("expected no outgoing metadata for empty baggage")
	}
}
