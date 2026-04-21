// Package ratelimiter provides token-bucket rate limiting for gRPC servers.
package ratelimiter

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor returns a gRPC unary server interceptor that enforces
// rate limiting using the provided Limiter. Requests that exceed the allowed
// rate are rejected with codes.ResourceExhausted.
func UnaryServerInterceptor(limiter *Limiter) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if !limiter.Allow() {
			return nil, status.Errorf(
				codes.ResourceExhausted,
				"rate limit exceeded for method %s",
				info.FullMethod,
			)
		}
		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a gRPC stream server interceptor that enforces
// rate limiting using the provided Limiter. Stream requests that exceed the
// allowed rate are rejected with codes.ResourceExhausted.
func StreamServerInterceptor(limiter *Limiter) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		if !limiter.Allow() {
			return status.Errorf(
				codes.ResourceExhausted,
				"rate limit exceeded for method %s",
				info.FullMethod,
			)
		}
		return handler(srv, ss)
	}
}
