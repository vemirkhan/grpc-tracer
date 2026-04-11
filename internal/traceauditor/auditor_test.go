package traceauditor_test

import (
	"testing"

	"github.com/user/grpc-tracer/internal/traceauditor"
)

func TestNew_EmptyAuditor(t *testing.T) {
	a := traceauditor.New()
	if a.Len() != 0 {
		t.Fatalf("expected 0 events, got %d", a.Len())
	}
}

func TestRecord_IncrementsLen(t *testing.T) {
	a := traceauditor.New()
	a.Record(traceauditor.EventWrite, "trace-1", "span-1", "svc-a", "added span")
	a.Record(traceauditor.EventRead, "trace-1", "span-1", "svc-b", "read span")
	if a.Len() != 2 {
		t.Fatalf("expected 2 events, got %d", a.Len())
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	a := traceauditor.New()
	a.Record(traceauditor.EventDelete, "trace-2", "span-2", "admin", "pruned")
	events := a.All()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Kind != traceauditor.EventDelete {
		t.Errorf("expected kind delete, got %s", events[0].Kind)
	}
	if events[0].Actor != "admin" {
		t.Errorf("expected actor admin, got %s", events[0].Actor)
	}
}

func TestFilterByTrace_ReturnsMatchingEvents(t *testing.T) {
	a := traceauditor.New()
	a.Record(traceauditor.EventWrite, "trace-A", "s1", "svc", "")
	a.Record(traceauditor.EventWrite, "trace-B", "s2", "svc", "")
	a.Record(traceauditor.EventRead, "trace-A", "s3", "svc", "")

	results := a.FilterByTrace("trace-A")
	if len(results) != 2 {
		t.Fatalf("expected 2 events for trace-A, got %d", len(results))
	}
	for _, e := range results {
		if e.TraceID != "trace-A" {
			t.Errorf("unexpected traceID %s", e.TraceID)
		}
	}
}

func TestFilterByTrace_NoMatch(t *testing.T) {
	a := traceauditor.New()
	a.Record(traceauditor.EventWrite, "trace-X", "s1", "svc", "")

	results := a.FilterByTrace("trace-MISSING")
	if len(results) != 0 {
		t.Fatalf("expected 0 events, got %d", len(results))
	}
}

func TestEvent_String_ContainsFields(t *testing.T) {
	a := traceauditor.New()
	a.Record(traceauditor.EventRead, "tid", "sid", "actor1", "some detail")
	events := a.All()
	s := events[0].String()
	for _, want := range []string{"read", "tid", "sid", "actor1", "some detail"} {
		if !contains(s, want) {
			t.Errorf("String() missing %q in %q", want, s)
		}
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		(func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		})())
}
