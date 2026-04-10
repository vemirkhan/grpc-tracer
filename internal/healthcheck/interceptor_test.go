package healthcheck

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var dummyInfo = &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

func successHandler context.Context, req interface{}) (interface{}, error) {
	return req, nil
}

func internalE _ interface{}) (interface{}, error) {
	return nil, status.Error(codes.Internal, "boom")
}

.Context, _ interface{}) (interface{}, error) {
	return nil, status.Error(codes.NotFound, "missing")
}

func plainErrHandler(_ context.Context, _ interface{}) (interface{}, error) {
	return nil, errors.New("plain error")
}

func TestInterceptor_SuccessMarksHealthy(t *testing.T) {
	c := New()
	intc := UnaryServerInterceptor(c, "svc")
	_, err := intc(context.Background(), nil, dummyInfo, successHandler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	h, ok := c.Get("svc")
	if !ok {
		t.Fatal("expected health record")
	}
	if h.Status != StatusHealthy {
		t.Errorf("expected healthy, got %s", h.Status)
	}
}

func TestInterceptor_InternalErrorMarksUnhealthy(t *testing.T) {
	c := New()
	intc := UnaryServerInterceptor(c, "svc")
	_, _ = intc(context.Background(), nil, dummyInfo, internalErrHandler)

	h, _ := c.Get("svc")
	if h.Status != StatusUnhealthy {
		t.Errorf("expected unhealthy, got %s", h.Status)
	}
}

func TestInterceptor_NotFoundMarksDegraded(t *testing.T) {
	c := New()
	intc := UnaryServerInterceptor(c, "svc")
	_, _ = intc(context.Background(), nil, dummyInfo, notFoundHandler)

	h, _ := c.Get("svc")
	if h.Status != StatusDegraded {
		t.Errorf("expected degraded, got %s", h.Status)
	}
}

func TestInterceptor_PlainErrorMarksDegraded(t *testing.T) {
	c := New()
	intc := UnaryServerInterceptor(c, "svc")
	_, _ = intc(context.Background(), nil, dummyInfo, plainErrHandler)

	h, _ := c.Get("svc")
	if h.Status != StatusDegraded {
		t.Errorf("expected degraded for unknown code, got %s", h.Status)
	}
}
