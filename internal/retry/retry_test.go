package retry_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/user/grpc-tracer/internal/retry"
)

func makeInvoker(responses []error) grpc.UnaryInvoker {
	i := 0
	return func(_ context.Context, _ string, _, _ interface{}, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
		if i >= len(responses) {
			return nil
		}
		er++
		return err
	}
}

func TestRetry_SuccessOnFirstAttempt(t *testing.T) {
	intc := retry.UnaryClientInterceptor(retry.Options{MaxAttempts: 3})
	invoker := makeInvoker([]error{nil})
	err := intc(context.Background(), "/svc/M", nil, nil, nil, invoker)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestRetry_RetriesOnUnavailable(t *testing.T) {
	unavail := status.Error(codes.Unavailable, "down")
	intc := retry.UnaryClientInterceptor(retry.Options{
		MaxAttempts: 3,
		Backoff:     1 * time.Millisecond,
	})
	invoker := makeInvoker([]error{unavail, unavail, nil})
	err := intc(context.Background(), "/svc/M", nil, nil, nil, invoker)
	if err != nil {
		t.Fatalf("expected success after retries, got %v", err)
	}
}

func TestRetry_ExceedsMaxAttempts(t *testing.T) {
	unavail := status.Error(codes.Unavailable, "down")
	intc := retry.UnaryClientInterceptor(retry.Options{
		MaxAttempts: 2,
		Backoff:     1 * time.Millisecond,
	})
	invoker := makeInvoker([]error{unavail, unavail, unavail})
	err := intc(context.Background(), "/svc/M", nil, nil, nil, invoker)
	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	if status.Code(err) != codes.Unavailable {
		t.Errorf("expected Unavailable, got %v", status.Code(err))
	}
}

func TestRetry_NonRetryableError(t *testing.T) {
	permErr := status.Error(codes.PermissionDenied, "no access")
	intc := retry.UnaryClientInterceptor(retry.Options{
		MaxAttempts: 3,
		Backoff:     1 * time.Millisecond,
	})
	calls := 0
	invoker := func(_ context.Context, _ string, _, _ interface{}, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
		calls++
		return permErr
	}
	err := intc(context.Background(), "/svc/M", nil, nil, nil, invoker)
	if !errors.Is(err, permErr) {
		t.Errorf("expected permErr, got %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestRetry_ContextCancelled(t *testing.T) {
	unavail := status.Error(codes.Unavailable, "down")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	intc := retry.UnaryClientInterceptor(retry.Options{
		MaxAttempts: 5,
		Backoff:     50 * time.Millisecond,
	})
	invoker := makeInvoker([]error{unavail, unavail, unavail, unavail, unavail})
	err := intc(ctx, "/svc/M", nil, nil, nil, invoker)
	if err == nil {
		t.Fatal("expected error due to cancelled context")
	}
}
