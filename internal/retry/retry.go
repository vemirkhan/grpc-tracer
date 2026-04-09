// Package retry provides a simple retry policy for gRPC unary calls.
// It wraps a gRPC UnaryInvoker and retries on transient errors up to
// a configurable maximum number of attempts with optional backoff.
package retry

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Options configures the retry behaviour.
type Options struct {
	// MaxAttempts is the total number of tries (including the first). Default: 3.
	MaxAttempts int
	// Backoff is the wait time between retries. Default: 100ms.
	Backoff time.t// ReteCodes lists gRPC status codes that trigger// Defaults to [Exhausted].
	RetryableCodes []codes.Code
}

func (o *Options) applyDefaults() {
	if o.MaxAttempts <= 0 {
		o.MaxAttempts = 3
	}
	if o.Backoff <= 0 {
		o.Backoff = 100 * time.Millisecond
	}
	if len(o.RetryableCodes) == 0 {
		o.RetryableCodes = []codes.Code{codes.Unavailable, codes.ResourceExhausted}
	}
}

func isRetryable(err error, codes []codes.Code) bool {
	st, ok := status.FromError(err)
	if !ok {
		return false
	}
	for _, c := range codes {
		if st.Code() == c {
			return true
		}
	}
	return false
}

// UnaryClientInterceptor returns a grpc.UnaryClientInterceptor that retries
// failed calls according to opts.
func UnaryClientInterceptor(opts Options) grpc.UnaryClientInterceptor {
	opts.applyDefaults()

	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		copts ...grpc.CallOption,
	) error {
		var lastErr error
		for attempt := 0; attempt < opts.MaxAttempts; attempt++ {
			if attempt > 0 {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(opts.Backoff):
				}
			}
			lastErr = invoker(ctx, method, req, reply, cc, copts...)
			if lastErr == nil {
				return nil
			}
			if !isRetryable(lastErr, opts.RetryableCodes) {
				return lastErr
			}
		}
		return lastErr
	}
}
