// Package middleware provides helpers for composing multiple gRPC unary server
// interceptors.
//
// # Chain
//
// Chain combines an arbitrary number of [grpc.UnaryServerInterceptor] values
// into a single interceptor. Interceptors execute in the order they are
// passed: the first interceptor wraps all subsequent ones, and the last
// interceptor is closest to the actual RPC handler.
//
// This is useful when you want to compose the interceptors provided by this
// module — for example, propagation, sampling, and tracing — without relying
// on the gRPC server's built-in chaining mechanism:
//
//	chain := middleware.Chain(
//		propagator.PropagatingUnaryServerInterceptor(store),
//		sampler.SampledUnaryServerInterceptor(s, store),
//		interceptor.UnaryServerInterceptor(store),
//	)
//
//	grpc.NewServer(grpc.UnaryInterceptor(chain))
package middleware
