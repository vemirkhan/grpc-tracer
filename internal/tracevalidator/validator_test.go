package tracevalidator_test

import (
	"testing"
	"time"

	"github.com/user/grpc-tracer/internal/storage"
	"github.com/user/grpc-tracer/internal/tracevalidator"
)

func makeSpan(traceID, spanID, service string, dur time.Duration, start time.Time) storage.Span {
	return storage.Span{
		TraceID:     traceID,
		SpanID:      spanID,
		ServiceName: service,
		Duration:    dur,
		StartTime:   start,
	}
}

func TestValidateSpan_Valid(t *testing.T) {
	v := tracevalidator.New(tracevalidator.Options{})
	s := makeSpan("trace1", "span1", "svc", 10*time.Millisecond, time.Now().Add(-time.Second))
	if err := v.ValidateSpan(s); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestValidateSpan_MissingTraceID(t *testing.T) {
	v := tracevalidator.New(tracevalidator.Options{})
	s := makeSpan("", "span1", "svc", 0, time.Now())
	if err := v.ValidateSpan(s); err != tracevalidator.ErrMissingTraceID {
		t.Fatalf("expected ErrMissingTraceID, got %v", err)
	}
}

func TestValidateSpan_MissingSpanID(t *testing.T) {
	v := tracevalidator.New(tracevalidator.Options{})
	s := makeSpan("trace1", "", "svc", 0, time.Now())
	if err := v.ValidateSpan(s); err != tracevalidator.ErrMissingSpanID {
		t.Fatalf("expected ErrMissingSpanID, got %v", err)
	}
}

func TestValidateSpan_MissingService(t *testing.T) {
	v := tracevalidator.New(tracevalidator.Options{})
	s := makeSpan("trace1", "span1", "  ", 0, time.Now())
	if err := v.ValidateSpan(s); err != tracevalidator.ErrMissingService {
		t.Fatalf("expected ErrMissingService, got %v", err)
	}
}

func TestValidateSpan_NegativeDuration(t *testing.T) {
	v := tracevalidator.New(tracevalidator.Options{})
	s := makeSpan("trace1", "span1", "svc", -1*time.Millisecond, time.Now())
	if err := v.ValidateSpan(s); err != tracevalidator.ErrNegativeDuration {
		t.Fatalf("expected ErrNegativeDuration, got %v", err)
	}
}

func TestValidateSpan_FutureStart(t *testing.T) {
	v := tracevalidator.New(tracevalidator.Options{})
	s := makeSpan("trace1", "span1", "svc", 0, time.Now().Add(time.Hour))
	if err := v.ValidateSpan(s); err != tracevalidator.ErrFutureStart {
		t.Fatalf("expected ErrFutureStart, got %v", err)
	}
}

func TestValidateSpan_AllowFutureStart(t *testing.T) {
	v := tracevalidator.New(tracevalidator.Options{AllowFutureStart: true})
	s := makeSpan("trace1", "span1", "svc", 0, time.Now().Add(time.Hour))
	if err := v.ValidateSpan(s); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestValidateSpan_MaxDurationExceeded(t *testing.T) {
	v := tracevalidator.New(tracevalidator.Options{MaxDuration: 100 * time.Millisecond})
	s := makeSpan("trace1", "span1", "svc", 200*time.Millisecond, time.Now())
	if err := v.ValidateSpan(s); err == nil {
		t.Fatal("expected error for exceeded max duration")
	}
}

func TestValidateAll_ReturnsAllErrors(t *testing.T) {
	v := tracevalidator.New(tracevalidator.Options{})
	spans := []storage.Span{
		makeSpan("trace1", "span1", "svc", 0, time.Now()),
		makeSpan("", "span2", "svc", 0, time.Now()),
		makeSpan("trace1", "", "svc", 0, time.Now()),
	}
	errs := v.ValidateAll(spans)
	if len(errs) != 2 {
		t.Fatalf("expected 2 errors, got %d", len(errs))
	}
	if errs[0] != nil {
		t.Errorf("span 0 should be valid")
	}
}
