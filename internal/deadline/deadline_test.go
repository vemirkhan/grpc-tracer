package deadline_test

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc"

	"github.com/user/grpc-tracer/internal/deadline"
)

var okHandler grpc.UnaryHandler = func(ctx context.Context, req interface{}) (interface{}, error) {
	return "ok", nil
}

func info() *grpc.UnaryServerInfo { return &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"} }

func TestNoConfig_PassesThrough(t *testing.T) {
	interceptor := deadline.UnaryServerInterceptor(deadline.Config{})
	resp, err := interceptor(context.Background(), nil, info(), okHandler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "ok" {
		t.Fatalf("expected 'ok', got %v", resp)
	}
}

func TestDefaultDeadline_InjectedWhenMissing(t *testing.T) {
	cfg := deadline.Config{DefaultDeadline: 5 * time.Second}
	interceptor := deadline.UnaryServerInterceptor(cfg)

	var capturedCtx context.Context
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		capturedCtx = ctx
		return "ok", nil
	}

	_, err := interceptor(context.Background(), nil, info(), handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	dl, ok := capturedCtx.Deadline()
	if !ok {
		t.Fatal("expected a deadline to be set")
	}
	if time.Until(dl) > 5*time.Second || time.Until(dl) <= 0 {
		t.Fatalf("deadline out of expected range: %v", dl)
	}
}

func TestMaxDeadline_CapsExistingDeadline(t *testing.T) {
	cfg := deadline.Config{MaxDeadline: 1 * time.Second}
	interceptor := deadline.UnaryServerInterceptor(cfg)

	// Provide a context with a very long deadline.
	ctxWithLong, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var capturedCtx context.Context
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		capturedCtx = ctx
		return "ok", nil
	}

	_, err := interceptor(ctxWithLong, nil, info(), handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	dl, ok := capturedCtx.Deadline()
	if !ok {
		t.Fatal("expected a deadline")
	}
	if time.Until(dl) > 1*time.Second+50*time.Millisecond {
		t.Fatalf("deadline was not capped: %v remaining", time.Until(dl))
	}
}

func TestMaxDeadline_DoesNotExtendShortDeadline(t *testing.T) {
	cfg := deadline.Config{MaxDeadline: 10 * time.Second}
	interceptor := deadline.UnaryServerInterceptor(cfg)

	ctxShort, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	var capturedCtx context.Context
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		capturedCtx = ctx
		return "ok", nil
	}

	_, _ = interceptor(ctxShort, nil, info(), handler)
	dl, ok := capturedCtx.Deadline()
	if !ok {
		t.Fatal("expected a deadline")
	}
	if time.Until(dl) > 200*time.Millisecond {
		t.Fatalf("short deadline was incorrectly extended: %v remaining", time.Until(dl))
	}
}
