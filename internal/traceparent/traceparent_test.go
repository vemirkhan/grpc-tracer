package traceparent_test

import (
	"testing"

	"github.com/your-org/grpc-tracer/internal/traceparent"
)

const (
	validTraceID  = "4bf92f3577b34da6a3ce929d0e0e4736"
	validParentID = "00f067aa0ba902b7"
)

func TestNew_SampledTrue(t *testing.T) {
	tp := traceparent.New(validTraceID, validParentID, true)
	if tp.Flags != "01" {
		t.Fatalf("expected flags=01, got %s", tp.Flags)
	}
	if !tp.Sampled() {
		t.Fatal("expected Sampled() == true")
	}
}

func TestNew_SampledFalse(t *testing.T) {
	tp := traceparent.New(validTraceID, validParentID, false)
	if tp.Sampled() {
		t.Fatal("expected Sampled() == false")
	}
}

func TestString_RoundTrip(t *testing.T) {
	tp := traceparent.New(validTraceID, validParentID, true)
	encoded := tp.String()
	parsed, err := traceparent.Parse(encoded)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if parsed.TraceID != tp.TraceID {
		t.Errorf("TraceID mismatch: got %s, want %s", parsed.TraceID, tp.TraceID)
	}
	if parsed.ParentID != tp.ParentID {
		t.Errorf("ParentID mismatch: got %s, want %s", parsed.ParentID, tp.ParentID)
	}
}

func TestParse_ValidHeader(t *testing.T) {
	header := "00-" + validTraceID + "-" + validParentID + "-01"
	tp, err := traceparent.Parse(header)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tp.TraceID != validTraceID {
		t.Errorf("TraceID: got %s, want %s", tp.TraceID, validTraceID)
	}
	if tp.Version != "00" {
		t.Errorf("Version: got %s, want 00", tp.Version)
	}
}

func TestParse_InvalidFormat(t *testing.T) {
	_, err := traceparent.Parse("not-a-traceparent")
	if err != traceparent.ErrInvalidFormat {
		t.Fatalf("expected ErrInvalidFormat, got %v", err)
	}
}

func TestParse_UnsupportedVersion(t *testing.T) {
	header := "ff-" + validTraceID + "-" + validParentID + "-00"
	_, err := traceparent.Parse(header)
	if err != traceparent.ErrUnsupportedVer {
		t.Fatalf("expected ErrUnsupportedVer, got %v", err)
	}
}

func TestParse_ZeroTraceID(t *testing.T) {
	zeroTrace := "00000000000000000000000000000000"
	header := "00-" + zeroTrace + "-" + validParentID + "-00"
	_, err := traceparent.Parse(header)
	if err != traceparent.ErrZeroTraceID {
		t.Fatalf("expected ErrZeroTraceID, got %v", err)
	}
}

func TestParse_ZeroParentID(t *testing.T) {
	zeroParent := "0000000000000000"
	header := "00-" + validTraceID + "-" + zeroParent + "-00"
	_, err := traceparent.Parse(header)
	if err != traceparent.ErrZeroParentID {
		t.Fatalf("expected ErrZeroParentID, got %v", err)
	}
}

func TestParse_CaseInsensitive(t *testing.T) {
	header := "00-" + "4BF92F3577B34DA6A3CE929D0E0E4736" + "-" + "00F067AA0BA902B7" + "-01"
	tp, err := traceparent.Parse(header)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tp.TraceID != validTraceID {
		t.Errorf("expected lowercase TraceID, got %s", tp.TraceID)
	}
}
