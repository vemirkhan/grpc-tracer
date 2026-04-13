package tracealert_test

import (
	"testing"
	"time"

	"github.com/user/grpc-tracer/internal/storage"
	"github.com/user/grpc-tracer/internal/tracealert"
)

func makeSpan(traceID, spanID, service string, dur time.Duration, errMsg string) storage.Span {
	return storage.Span{
		TraceID:  traceID,
		SpanID:   spanID,
		Service:  service,
		Duration: dur,
		Error:    errMsg,
	}
}

func TestNew_Defaults(t *testing.T) {
	a := tracealert.New(tracealert.Config{})
	if a == nil {
		t.Fatal("expected non-nil alerter")
	}
}

func TestEvaluate_NoAlert_FastHealthySpan(t *testing.T) {
	a := tracealert.New(tracealert.Config{})
	span := makeSpan("t1", "s1", "svc", 10*time.Millisecond, "")
	a.Evaluate(span)
	if got := a.All(); len(got) != 0 {
		t.Fatalf("expected no alerts, got %d", len(got))
	}
}

func TestEvaluate_WarnOnSlowSpan(t *testing.T) {
	a := tracealert.New(tracealert.Config{
		WarnDuration:  100 * time.Millisecond,
		ErrorDuration: 1 * time.Second,
	})
	span := makeSpan("t1", "s1", "svc", 200*time.Millisecond, "")
	a.Evaluate(span)
	alerts := a.All()
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}
	if alerts[0].Level != tracealert.LevelWarn {
		t.Errorf("expected warn level, got %s", alerts[0].Level)
	}
}

func TestEvaluate_ErrorOnVerySlowSpan(t *testing.T) {
	a := tracealert.New(tracealert.Config{
		WarnDuration:  100 * time.Millisecond,
		ErrorDuration: 500 * time.Millisecond,
	})
	span := makeSpan("t1", "s1", "svc", 600*time.Millisecond, "")
	a.Evaluate(span)
	alerts := a.All()
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}
	if alerts[0].Level != tracealert.LevelError {
		t.Errorf("expected error level, got %s", alerts[0].Level)
	}
}

func TestEvaluate_AlertOnError(t *testing.T) {
	a := tracealert.New(tracealert.Config{AlertOnError: true})
	span := makeSpan("t1", "s1", "svc", 1*time.Millisecond, "rpc failed")
	a.Evaluate(span)
	alerts := a.All()
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}
	if alerts[0].Level != tracealert.LevelError {
		t.Errorf("expected error level, got %s", alerts[0].Level)
	}
}

func TestEvaluate_NoAlertOnError_WhenDisabled(t *testing.T) {
	a := tracealert.New(tracealert.Config{AlertOnError: false})
	span := makeSpan("t1", "s1", "svc", 1*time.Millisecond, "some error")
	a.Evaluate(span)
	if got := a.All(); len(got) != 0 {
		t.Fatalf("expected no alerts, got %d", len(got))
	}
}

func TestClear_RemovesAlerts(t *testing.T) {
	a := tracealert.New(tracealert.Config{AlertOnError: true})
	a.Evaluate(makeSpan("t1", "s1", "svc", 1*time.Millisecond, "err"))
	if len(a.All()) == 0 {
		t.Fatal("expected alerts before clear")
	}
	a.Clear()
	if got := a.All(); len(got) != 0 {
		t.Fatalf("expected 0 alerts after clear, got %d", len(got))
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	a := tracealert.New(tracealert.Config{AlertOnError: true})
	a.Evaluate(makeSpan("t1", "s1", "svc", 1*time.Millisecond, "err"))
	first := a.All()
	first[0].Service = "mutated"
	second := a.All()
	if second[0].Service == "mutated" {
		t.Error("All() should return a copy, not a reference")
	}
}
