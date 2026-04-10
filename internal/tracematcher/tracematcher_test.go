package tracematcher_test

import (
	"testing"
	"time"

	"github.com/user/grpc-tracer/internal/storage"
	"github.com/user/grpc-tracer/internal/tracematcher"
)

func makeTrace(traceID, service, method, errMsg string, dur time.Duration, tags map[string]string) storage.Trace {
	span := storage.Span{
		TraceID:     traceID,
		SpanID:      "span-1",
		ServiceName: service,
		Method:      method,
		Error:       errMsg,
		Duration:    dur,
		StartTime:   time.Now(),
		Tags:        tags,
	}
	return storage.Trace{
		TraceID: traceID,
		Spans:   []storage.Span{span},
	}
}

func TestMatch_NoCriteria_AlwaysTrue(t *testing.T) {
	tr := makeTrace("t1", "svc", "/Hello", "", 10*time.Millisecond, nil)
	if !tracematcher.Match(tr, tracematcher.Criteria{}) {
		t.Fatal("empty criteria should match any trace")
	}
}

func TestMatch_ByTraceID(t *testing.T) {
	tr := makeTrace("abc123", "svc", "/Hello", "", 0, nil)
	if !tracematcher.Match(tr, tracematcher.Criteria{TraceID: "abc123"}) {
		t.Fatal("expected match on exact trace ID")
	}
	if tracematcher.Match(tr, tracematcher.Criteria{TraceID: "other"}) {
		t.Fatal("should not match different trace ID")
	}
}

func TestMatch_ByServiceName(t *testing.T) {
	tr := makeTrace("t1", "OrderService", "/Place", "", 0, nil)
	if !tracematcher.Match(tr, tracematcher.Criteria{ServiceName: "order"}) {
		t.Fatal("expected case-insensitive service match")
	}
	if tracematcher.Match(tr, tracematcher.Criteria{ServiceName: "payment"}) {
		t.Fatal("should not match unrelated service")
	}
}

func TestMatch_OnlyErrors(t *testing.T) {
	errTrace := makeTrace("t2", "svc", "/Op", "rpc error", 0, nil)
	okTrace := makeTrace("t3", "svc", "/Op", "", 0, nil)

	if !tracematcher.Match(errTrace, tracematcher.Criteria{OnlyErrors: true}) {
		t.Fatal("error trace should match OnlyErrors")
	}
	if tracematcher.Match(okTrace, tracematcher.Criteria{OnlyErrors: true}) {
		t.Fatal("ok trace should not match OnlyErrors")
	}
}

func TestMatch_ByMinDuration(t *testing.T) {
	fast := makeTrace("t4", "svc", "/Op", "", 5*time.Millisecond, nil)
	slow := makeTrace("t5", "svc", "/Op", "", 200*time.Millisecond, nil)

	c := tracematcher.Criteria{MinDuration: 100 * time.Millisecond}
	if tracematcher.Match(fast, c) {
		t.Fatal("fast trace should not match min duration")
	}
	if !tracematcher.Match(slow, c) {
		t.Fatal("slow trace should match min duration")
	}
}

func TestMatch_ByTag(t *testing.T) {
	tr := makeTrace("t6", "svc", "/Op", "", 0, map[string]string{"env": "prod"})
	if !tracematcher.Match(tr, tracematcher.Criteria{TagKey: "env", TagValue: "prod"}) {
		t.Fatal("expected tag match")
	}
	if tracematcher.Match(tr, tracematcher.Criteria{TagKey: "env", TagValue: "staging"}) {
		t.Fatal("wrong tag value should not match")
	}
}

func TestMatchAll_FiltersStore(t *testing.T) {
	store := storage.NewTraceStore()
	errSpan := storage.Span{TraceID: "e1", SpanID: "s1", ServiceName: "svc", Error: "boom", StartTime: time.Now()}
	okSpan := storage.Span{TraceID: "o1", SpanID: "s2", ServiceName: "svc", StartTime: time.Now()}
	store.AddSpan(errSpan)
	store.AddSpan(okSpan)

	results := tracematcher.MatchAll(store, tracematcher.Criteria{OnlyErrors: true})
	if len(results) != 1 {
		t.Fatalf("expected 1 error trace, got %d", len(results))
	}
	if results[0].TraceID != "e1" {
		t.Errorf("expected trace e1, got %s", results[0].TraceID)
	}
}
