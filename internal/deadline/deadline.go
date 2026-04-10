// Package deadline provides utilities for propagating and enforcing
// per-RPC deadlines across gRPC call chains.
package deadline

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Config holds configuration for deadline enforcement.
type Config struct {
// MaxDeadline is the upper bound applied to any RPC deadline.
	// If zero, no upper bound is enforced.
	MaxDeadline time.Duration

	// DefaultDeadline is applied when the incoming context has no deadline.
	// If zero, no default is injected.
	DefaultDeadline time.Duration
}

// UnaryServerInterceptor returns a gRPC server interceptor that enforces
// deadline constraints defined by cfg. It caps incoming deadlines at
// MaxDeadline and injects DefaultDeadline when none is present.
func UnaryServerInterceptor(cfg Config) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		_ *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		ctx, cancel := applyDeadline(ctx, cfg)
		defer cancel()

		resp, err := handler(ctx, req)
		if err != nil {
			return nil, err
		}
		if ctx.Err() == context.DeadlineExceeded {
			return nil, status.Error(codes.DeadlineExceeded, "deadline exceeded")
		}
		return resp, nil
	}
}

// applyDeadline adjusts ctx according to cfg and returns the new context
// together with its cancel function.
func applyDeadline(ctx context.Context, cfg Config) (context.Context, context.CancelFunc) {
	now := time.Now()

	// Determine the effective deadline for this call.
	var target time.Time

	if dl, ok := ctx.Deadline(); ok {
		target = dl
	} else if cfg.DefaultDeadline > 0 {
		target = now.Add(cfg.DefaultDeadline)
	}

	// Cap at MaxDeadline if configured.
	if cfg.MaxDeadline > 0 {
		cap := now.Add(cfg.MaxDeadline)
		if target.IsZero() || cap.Before(target) {
			target = cap
		}
	}

	if target.IsZero() {
		return ctx, func() {}
	}
	return context.WithDeadline(ctx, target)
}
