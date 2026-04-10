// Package logger provides structured span-level logging for gRPC trace events.
package logger

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/user/grpc-tracer/internal/storage"
)

// Level represents the logging verbosity.
type Level int

const (
	LevelInfo  Level = iota
	LevelWarn
	LevelError
)

func (l Level) String() string {
	switch l {
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "INFO"
	}
}

// Logger writes structured trace log lines to an io.Writer.
type Logger struct {
	out   io.Writer
	level Level
}

// New creates a Logger writing to out at the given minimum level.
// If out is nil, os.Stdout is used.
func New(out io.Writer, level Level) *Logger {
	if out == nil {
		out = os.Stdout
	}
	return &Logger{out: out, level: level}
}

// LogSpan emits a single log line for a span if its level meets the threshold.
func (l *Logger) LogSpan(span storage.Span) {
	var lvl Level
	if span.Error != "" {
		lvl = LevelError
	} else if span.Duration > 500*time.Millisecond {
		lvl = LevelWarn
	} else {
		lvl = LevelInfo
	}

	if lvl < l.level {
		return
	}

	fmt.Fprintf(
		l.out,
		"[%s] trace=%s span=%s service=%s method=%s duration=%s error=%q\n",
		lvl,
		span.TraceID,
		span.SpanID,
		span.ServiceName,
		span.Method,
		span.Duration,
		span.Error,
	)
}

// LogTrace emits log lines for every span in a trace.
func (l *Logger) LogTrace(spans []storage.Span) {
	for _, s := range spans {
		l.LogSpan(s)
	}
}
