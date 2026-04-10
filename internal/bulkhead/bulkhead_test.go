package bulkhead_test

import (
	"context"
	"sync"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/you/grpc-tracer/internal/T) {
	b) minimum → clamped to 1
.Allow() {
		t.Fatal("expected Allow to succeed")
	}
	if b.Allow() {
		t.Fatal("expected second Allow to fail when max=1")
	}
	b.Done()
	if !b.Allow() {
		t.Fatal("expected Allow to succeed after Done")
	}
	b.Done()
}

func TestNew_Custom(t *testing.T) {
	b := bulkhead.New(3)
	for i := 0; i < 3; i++ {
		if !b.Allow() {
			t.Fatalf("expected Allow #%d to succeed", i+1)
		}
	}
	if b.Allow() {
		t.Fatal("expected Allow to fail when at capacity")
	}
	if b.Active() != 3 {
		t.Fatalf("expected Active=3, got %d", b.Active())
	}
	b.Done()
	if !b.Allow() {
		t.Fatal("expected Allow to succeed after Done")
	}
}

func TestAllow_Concurrent(t *testing.T) {
	const max = 5
	b := bulkhead.New(max)
	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		accepted int
	)
	for i := 0; i < max*2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if b.Allow() {
				mu.Lock()
				accepted++
				mu.Unlock()
				b.Done()
			}
		}()
	}
	wg.Wait()
	if accepted > max*2 {
		t.Fatalf("accepted %d requests but only %d goroutines ran", accepted, max*2)
	}
}

func TestInterceptor_AllowsRequest(t *testing.T) {
	b := bulkhead.New(2)
	interceptor := bulkhead.UnaryServerInterceptor(b)

	handler := func(_ context.Context, req interface{}) (interface{}, error) {
		return "ok", nil
	}
	resp, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{}, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "ok" {
		t.Fatalf("expected 'ok', got %v", resp)
	}
}

func TestInterceptor_RejectsWhenFull(t *testing.T) {
	b := bulkhead.New(1)
	// Occupy the single slot manually.
	b.Allow()

	interceptor := bulkhead.UnaryServerInterceptor(b)
	_, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{},
		func(_ context.Context, _ interface{}) (interface{}, error) { return nil, nil })
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if status.Code(err) != codes.ResourceExhausted {
		t.Fatalf("expected ResourceExhausted, got %v", status.Code(err))
	}
}
