// Package ratelimiter implements a token-bucket rate limiter for use with
// gRPC servers.
//
// # Overview
//
// The Limiter type manages a token bucket that refills at a configurable rate
// (tokens per second) up to a configurable burst size. Each call to Allow
// consumes one token; if no tokens are available the call returns false.
//
// # gRPC Integration
//
// UnaryServerInterceptor wraps a Limiter and returns a grpc.UnaryServerInterceptor
// that rejects excess requests with codes.ResourceExhausted, allowing the
// caller to apply back-pressure without crashing the server.
//
// # Example
//
//	limiter := ratelimiter.New(ratelimiter.Config{Rate: 100, Burst: 20})
//	server := grpc.NewServer(
//		grpc.UnaryInterceptor(ratelimiter.UnaryServerInterceptor(limiter)),
//	)
package ratelimiter
