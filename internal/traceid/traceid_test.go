package traceid_test

import (
	"strings"
	"testing"

	"github.com/your-org/grpc-tracer/internal/traceid"
)

func TestNewTraceID_Length(t *testing.T) {
	id, err := traceid.NewTraceID()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 16 bytes → 32 hex chars
	if got := len(id); got != 32 {
		t.Errorf("expected length 32, got %d", got)
	}
}

func TestNewSpanID_Length(t *testing.T) {
	id, err := traceid.NewSpanID()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 8 bytes → 16 hex chars
	if got := len(id); got != 16 {
		t.Errorf("expected length 16, got %d", got)
	}
}

func TestNewTraceID_Uniqueness(t *testing.T) {
	seen := make(map[string]struct{}, 100)
	for i := 0; i < 100; i++ {
		id, err := traceid.NewTraceID()
		if err != nil {
			t.Fatalf("iteration %d: unexpected error: %v", i, err)
		}
		if _, dup := seen[id]; dup {
			t.Fatalf("duplicate trace ID generated: %s", id)
		}
		seen[id] = struct{}{}
	}
}

func TestNewTraceID_IsLowerHex(t *testing.T) {
	id, _ := traceid.NewTraceID()
	if id != strings.ToLower(id) {
		t.Errorf("trace ID is not lowercase: %s", id)
	}
}

func TestMustTraceID_DoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("MustTraceID panicked: %v", r)
		}
	}()
	id := traceid.MustTraceID()
	if id == "" {
		t.Error("expected non-empty ID")
	}
}

func TestMustSpanID_DoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("MustSpanID panicked: %v", r)
		}
	}()
	id := traceid.MustSpanID()
	if id == "" {
		t.Error("expected non-empty ID")
	}
}

func TestValidate(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid trace id", "4bf92f3577b34da6a3ce929d0e0e4736", true},
		{"valid span id", "00f067aa0ba902b7", true},
		{"uppercase accepted", "4BF92F3577B34DA6", true},
		{"empty string", "", false},
		{"odd length", "abc", false},
		{"non-hex chars", "zzzzzzzzzzzzzzzz", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := traceid.Validate(tc.input); got != tc.want {
				t.Errorf("Validate(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}
