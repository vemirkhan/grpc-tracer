package timeout_test

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/your-org/grpc-tracer/internal/timeout"
)

// slowInvoker simulates an RPC that blocks until string, _, _ interface{}, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
		select {
		case <-timefunc TestUnaryClientInterceptor_DefaultTimeout(t *testing.T) {
	intercept := timeout.UnaryClientInterceptor(timeout.Options{})
	// A fast invoker should succeed even with the default timeout.
	invoker := slowInvoker(0)
	err := intercept(context.Background(), "/svc/Method", nil, nil, nil, invoker)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestUnaryClientInterceptor_TimesOut(t *testing.T) {
	intercept := timeout.UnaryClientInterceptor(timeout.Options{Timeout: 20 * time.Millisecond})
	// Invoker takes longer than the configured timeout.
	invoker := slowInvoker(500 * time.Millisecond)

	err := intercept(context.Background(), "/svc/Slow", nil, nil, nil, invoker)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %T: %v", err, err)
	}
	if st.Code() != codes.DeadlineExceeded {
		t.Errorf("expected DeadlineExceeded, got %v", st.Code())
	}
}

func TestUnaryClientInterceptor_RespectsExistingDeadline(t *testing.T) {
	// Configure a generous interceptor timeout.
	intercept := timeout.UnaryClientInterceptor(timeout.Options{Timeout: 10 * time.Second})

	// Parent context has a much shorter deadline.
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	invoker := slowInvoker(500 * time.Millisecond)
	err := intercept(ctx, "/svc/Method", nil, nil, nil, invoker)
	if err == nil {
		t.Fatal("expected context deadline error, got nil")
	}
}

func TestUnaryClientInterceptor_FastSuccess(t *testing.T) {
	intercept := timeout.UnaryClientInterceptor(timeout.Options{Timeout: 100 * time.Millisecond})
	var called bool
	invoker := func(_ context.Context, _ string, _, _ interface{}, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
		called = true
		return nil
	}
	err := intercept(context.Background(), "/svc/Fast", nil, nil, nil, invoker)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("invoker was never called")
	}
}
