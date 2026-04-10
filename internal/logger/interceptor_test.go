package logger_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/user/grpc-tracer/internal/logger"
	"google.golang.org/grpc"
)

func okHandler(_ context.Context, _ interface{}) (interface{}, error) {
	return "ok", nil
}

func errHandler(_ context.Context, _ interface{}) (interface{}, error) {
	return nil, errors.New("boom")
}

func TestInterceptor_LogsSuccessfulCall(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(&buf, logger.LevelInfo)
	intercept := logger.UnaryServerInterceptor(l, "test-service")

	info := &grpc.UnaryServerInfo{FullMethod: "/pkg.Svc/Hello"}
	_, err := intercept(context.Background(), nil, info, okHandler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "test-service") {
		t.Errorf("expected service name in log, got: %s", out)
	}
	if !strings.Contains(out, "/pkg.Svc/Hello") {
		t.Errorf("expected method in log, got: %s", out)
	}
}

func TestInterceptor_LogsErrorCall(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(&buf, logger.LevelInfo)
	intercept := logger.UnaryServerInterceptor(l, "err-service")

	info := &grpc.UnaryServerInfo{FullMethod: "/pkg.Svc/Fail"}
	_, err := intercept(context.Background(), nil, info, errHandler)
	if err == nil {
		t.Fatal("expected error")
	}
	out := buf.String()
	if !strings.Contains(out, "ERROR") {
		t.Errorf("expected ERROR level for failed handler, got: %s", out)
	}
	if !strings.Contains(out, "boom") {
		t.Errorf("expected error message in log, got: %s", out)
	}
}

func TestInterceptor_PassesThroughResponse(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(&buf, logger.LevelInfo)
	intercept := logger.UnaryServerInterceptor(l, "svc")

	info := &grpc.UnaryServerInfo{FullMethod: "/pkg.Svc/Echo"}
	resp, err := intercept(context.Background(), "input", info, okHandler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "ok" {
		t.Errorf("expected 'ok', got %v", resp)
	}
}
