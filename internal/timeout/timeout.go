// Package timeout provides a gRPC unary client interceptor that enforces
// a configurable deadline on outgoing RPC calls.
package timeout

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DefaultTimeout is used when no timeout is explicitly configured.
const DefaultTimeout = 5 * time.Second

// Options holds configuration for the timeout interceptor.
type Options struct {
	// Timeout is the maximum duration allowed for a single RPC.
	// If zero, DefaultTimeout is used.
	Timeout time.Duration
}

// UnaryClientInterceptor returns a grpc.UnaryClientInterceptor that cancels
// the RPC if it exceeds the configured timeout. If the parent context already
// carries a shorter deadline, that deadline takes precedence.
func UnaryClientInterceptor(opts Options) grpc.UnaryClientInterceptor {
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		callOpts ...grpc.CallOption,
	) error {
		// Only apply our deadline if the context doesn't already have one
		// that expires sooner.
		if deadline, ok := ctx.Deadline(); !ok || time.Until(deadline) > timeout {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}

		err := invoker(ctx, method, req, reply, cc, callOpts...)
		if err != nil {
			// Translate DeadlineExceeded from context cancellation into a
			// gRPC DeadlineExceeded status so callers get a consistent code.
			if ctx.Err() == context.DeadlineExceeded {
				return status.Errorf(codes.DeadlineExceeded,
					"rpc timed out after %s: %s", timeout, method)
			}
		}
		return err
	}
}
