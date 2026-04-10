// Package middleware integration test: verifies that a rate-limiting
// interceptor composes correctly inside a middleware chain.
package middleware_test

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/user/grpc-tracer/internal/middleware"
	"github.com/user/grpc-tracer/internal/ratelimiter"
)

func TestChain_WithRateLimAllows(t *testing.T) {
	limiter := r100, Burst: 10})
	rlInterceptor := ratelimiter.UnaryServerInterceptor(limiter)

	chained := middleware.Chain(rlInterceptor)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "pong", nil
	}
	info := &grpc.UnaryServerInfo{FullMethod: "/svc/Ping"}

	resp, err := chained(context.Background(), "ping", info, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "pong" {
		t.Fatalf("expected pong, got %v", resp)
	}
}

func TestChain_WithRateLimiter_Rejects(t *testing.T) {
	// Burst of 0 is clamped to 1 internally; exhaust it first.
	limiter := ratelimiter.New(ratelimiter.Config{Rate: 1, Burst: 1})
	rlInterceptor := ratelimiter.UnaryServerInterceptor(limiter)

	chained := middleware.Chain(rlInterceptor)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "pong", nil
	}
	info := &grpc.UnaryServerInfo{FullMethod: "/svc/Ping"}

	// First call consumes the token.
	_, _ = chained(context.Background(), nil, info, handler)

	// Second call should be rate-limited.
	_, err := chained(context.Background(), nil, info, handler)
	if err == nil {
		t.Fatal("expected rate-limit error")
	}
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.ResourceExhausted {
		t.Fatalf("expected ResourceExhausted, got %v", err)
	}
}
