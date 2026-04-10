package logger_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/user/grpc-tracer/internal/logger"
	"github.com/user/grpc-tracer/internal/storage"
)

func makeSpan(svc, method, errMsg string, dur time.Duration) storage.Span {
	return storage.Span{
		TraceID:     "trace-1",
		SpanID:      "span-1",
		ServiceName: svc,
		Method:      method,
		Duration:    dur,
		Error:       errMsg,
	}
}

func TestLogSpan_InfoLevel(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(&buf, logger.LevelInfo)
	s := makeSpan("svc", "/pkg.Svc/Method", "", 10*time.Millisecond)
	l.LogSpan(s)
	out := buf.String()
	if !strings.Contains(out, "INFO") {
		t.Errorf("expected INFO in output, got: %s", out)
	}
}

func TestLogSpan_WarnOnSlowSpan(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(&buf, logger.LevelInfo)
	s := makeSpan("svc", "/pkg.Svc/Slow", "", 600*time.Millisecond)
	l.LogSpan(s)
	out := buf.String()
	if !strings.Contains(out, "WARN") {
		t.Errorf("expected WARN for slow span, got: %s", out)
	}
}

func TestLogSpan_ErrorOnFailedSpan(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(&buf, logger.LevelInfo)
	s := makeSpan("svc", "/pkg.Svc/Fail", "rpc error", 5*time.Millisecond)
	l.LogSpan(s)
	out := buf.String()
	if !strings.Contains(out, "ERROR") {
		t.Errorf("expected ERROR for failed span, got: %s", out)
	}
}

func TestLogSpan_FilteredByLevel(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(&buf, logger.LevelError)
	s := makeSpan("svc", "/pkg.Svc/Ok", "", 10*time.Millisecond)
	l.LogSpan(s)
	if buf.Len() != 0 {
		t.Errorf("expected no output for INFO span at ERROR level, got: %s", buf.String())
	}
}

func TestLogTrace_MultipleSpans(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(&buf, logger.LevelInfo)
	spans := []storage.Span{
		makeSpan("a", "/A/M", "", 1*time.Millisecond),
		makeSpan("b", "/B/M", "err", 2*time.Millisecond),
	}
	l.LogTrace(spans)
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 log lines, got %d", len(lines))
	}
}

func TestNew_NilWriterDefaultsToStdout(t *testing.T) {
	// Should not panic when out is nil.
	l := logger.New(nil, logger.LevelInfo)
	if l == nil {
		t.Fatal("expected non-nil logger")
	}
}
