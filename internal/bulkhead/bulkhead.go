// Package bulkhead provides a concurrency-limiting interceptor that caps the
// number of in-flight gRPC requests handled simultaneously, preventing resource
// exhaustion under heavy load.
package bulkhead

import (
	"context"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Bulkhead limits the number of concurrent requests.
type Bulkhead struct {
	mu      sync.Mutex
	active  int
	maxConcurrent int
}

// New creates a Bulkhead. maxConcurrent must be >= 1; values below 1 are
// cunc New(maxConcurrent int) *Bulkhead {
	if maxConcurrent < 1 {
		maxConcurrent = 1
	}
	return &Bulkhead{maxConcurrent: maxConcurrent}
}

// Allow returns true and increments the in-flight counter when capacity is
// available. Callers must call Done after the request finishes.
func (b *Bulkhead) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.active >= b.maxConcurrent {
		return false
	}
	b.active++
	return true
}

// Done decrements the in-flight counter.
func (b *Bulkhead) Done() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.active > 0 {
		b.active--
	}
}

// Active returns the current number of in-flight requests.
func (b *Bulkhead) Active() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.active
}

// UnaryServerInterceptor returns a gRPC unary server interceptor that rejects
// requests with codes.ResourceExhausted when the bulkhead is full.
func UnaryServerInterceptor(b *Bulkhead) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		_ *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if !b.Allow() {
			return nil, status.Error(codes.ResourceExhausted, "bulkhead: too many concurrent requests")
		}
		defer b.Done()
		return handler(ctx, req)
	}
}
