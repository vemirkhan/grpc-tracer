package middleware_test

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc"

	"github.com/you/grpc-tracer/internal/middleware"
)

var dummyInfo = &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

func echoHandler(_ context.Context, req interface{}) (interface{}, error) {
	return req, nil
}

func orderRecorder(tag string, log *[]string) middleware.UnaryInterceptorFunc {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		*log = append(*log, tag+":before")
		resp, err := handler(ctx, req)
		*log = append(*log, tag+":after")
		return resp, err
	}
}

func TestChain_Empty(t *testing.T) {
	chain := middleware.Chain()
	resp, err := chain(context.Background(), "ping", dummyInfo, echoHandler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "ping" {
		t.Fatalf("expected ping, got %v", resp)
	}
}

func TestChain_Single(t *testing.T) {
	var log []string
	chain := middleware.Chain(orderRecorder("A", &log))
	_, _ = chain(context.Background(), "x", dummyInfo, echoHandler)
	if len(log) != 2 || log[0] != "A:before" || log[1] != "A:after" {
		t.Fatalf("unexpected log: %v", log)
	}
}

func TestChain_Order(t *testing.T) {
	var log []string
	chain := middleware.Chain(
		orderRecorder("A", &log),
		orderRecorder("B", &log),
		orderRecorder("C", &log),
	)
	_, _ = chain(context.Background(), "x", dummyInfo, echoHandler)

	want := []string{"A:before", "B:before", "C:before", "C:after", "B:after", "A:after"}
	if len(log) != len(want) {
		t.Fatalf("log length mismatch: got %v", log)
	}
	for i, v := range want {
		if log[i] != v {
			t.Errorf("log[%d]: want %q, got %q", i, v, log[i])
		}
	}
}

func TestChain_PropagatesError(t *testing.T) {
	errBoom := errors.New("boom")
	errInterceptor := func(
		ctx context.Context,
		req interface{},
		_ *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		return nil, errBoom
	}

	chain := middleware.Chain(orderRecorder("A", &[]string{}), errInterceptor)
	_, err := chain(context.Background(), "x", dummyInfo, echoHandler)
	if !errors.Is(err, errBoom) {
		t.Fatalf("expected errBoom, got %v", err)
	}
}
