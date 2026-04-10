package ratelimiter

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	okHandler grpc.UnaryHandler =n	info = &grpc.UnaryServerInfoFullMethod: "/test/Method"}
)

func TestInterceptor_AllowsRequestWhenTokensAvailable(t *testing.T) {
	limiter := New(Config{Rate: 10, Burst: 5})
	intercept := UnaryServerInterceptor(limiter)

	resp, err := intercept(context.Background(), nil, info, okHandler)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp != "ok" {
		t.Fatalf("expected 'ok', got %v", resp)
	}
}

func TestInterceptor_RejectsRequestWhenExhausted(t *testing.T) {
	// Burst of 1 means only 1 token available initially.
	limiter := New(Config{Rate: 1, Burst: 1})
	intercept := UnaryServerInterceptor(limiter)

	// Consume the only available token.
	_, _ = intercept(context.Background(), nil, info, okHandler)

	// Second request should be rejected.
	_, err := intercept(context.Background(), nil, info, okHandler)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %v", err)
	}
	if st.Code() != codes.ResourceExhausted {
		t.Fatalf("expected ResourceExhausted, got %v", st.Code())
	}
}

func TestInterceptor_AllowsAfterRefill(t *testing.T) {
	// Rate of 100/s with burst 1: refills quickly.
	limiter := New(Config{Rate: 100, Burst: 1})
	intercept := UnaryServerInterceptor(limiter)

	// Exhaust the token.
	_, _ = intercept(context.Background(), nil, info, okHandler)

	// Wait for refill.
	time.Sleep(20 * time.Millisecond)

	_, err := intercept(context.Background(), nil, info, okHandler)
	if err != nil {
		t.Fatalf("expected request to be allowed after refill, got %v", err)
	}
}

func TestInterceptor_PropagatesHandlerError(t *testing.T) {
	limiter := New(Config{Rate: 10, Burst: 5})
	intercept := UnaryServerInterceptor(limiter)

	errHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, status.Error(codes.Internal, "handler error")
	}

	_, err := intercept(context.Background(), nil, info, errHandler)
	if err == nil {
		t.Fatal("expected handler error to propagate")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.Internal {
		t.Fatalf("expected Internal, got %v", st.Code())
	}
}
