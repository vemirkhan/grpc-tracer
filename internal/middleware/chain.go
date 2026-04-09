// Package middleware provides utilities for chaining multiple gRPC
// unary server interceptors into a single interceptor.
package middleware

import (
	"context"

	"google.golang.org/grpc"
)

// UnaryInterceptorFunc is an alias for grpc.UnaryServerInterceptor.
type UnaryInterceptorFunc = grpc.UnaryServerInterceptor

// Chain combines multiple UnaryServerInterceptors into one. Interceptors are
// applied in the order they are provided: the first interceptor is the
// outermost wrapper and the last is closest to the handler.
//
// Example:
//
//	chain := middleware.Chain(
//		propagatingInterceptor,
//		sampledInterceptor,
//		loggingInterceptor,
//	)
func Chain(interceptors ...UnaryInterceptorFunc) UnaryInterceptorFunc {
	switch len(interceptors) {
	case 0:
		return passthrough
	case 1:
		return interceptors[0]
	}

	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		return interceptors[0](ctx, req, info, chainHandler(interceptors[1:], info, handler))
	}
}

// chainHandler recursively builds a handler that calls each interceptor in
// sequence, eventually invoking the real handler.
func chainHandler(
	interceptors []UnaryInterceptorFunc,
	info *grpc.UnaryServerInfo,
	final grpc.UnaryHandler,
) grpc.UnaryHandler {
	if len(interceptors) == 0 {
		return final
	}
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return interceptors[0](ctx, req, info, chainHandler(interceptors[1:], info, final))
	}
}

// passthrough is a no-op interceptor used when Chain receives no arguments.
func passthrough(
	ctx context.Context,
	req interface{},
	_ *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	return handler(ctx, req)
}
