package filter_test

import (
	"testing"
	"time"

	"github.com/user/grpc-tracer/internal/filter"
	"github.com/user/grpc-tracer/internal/storage"
)

func makeSpan(traceID, service, errMsg string, dur time.Duration) storage.Span {
	now := time.Now()
	return storage.Span{
		TraceID:     traceID,
		SpanID:      "span-1",
		ServiceName: service,
		Method:      "/svc/Method",
		StartTime:   now,
		EndTime:     now.Add(dur),
		Error:       errMsg,
	}
}

func TestFilter_ByServiceName(t *testing.T) {
	spans := []storage.Span{
		makeSpan("t1", "auth", "", 10*time.Millisecond),
		makeSpan("t2", "payment", "", 20*time.Millisecond),
	}
	got := filter.Filter(spans, filter.Criteria{ServiceName: "auth"})
	if len(got) != 1 || got[0].ServiceName != "auth" {
		t.Fatalf("expected 1 auth span, got %v", got)
	}
}

func TestFilter_OnlyErrors(t *testing.T) {
	spans := []storage.Span{
		makeSpan("t1", "svc", "", 5*time.Millisecond),
		makeSpan("t2", "svc", "rpc error", 5*time.Millisecond),
	}
	got := filter.Filter(spans, filter.Criteria{OnlyErrors: true})
	if len(got) != 1 || got[0].Error == "" {
		t.Fatalf("expected 1 error span, got %v", got)
	}
}

func TestFilter_ByDuration(t *testing.T) {
	spans := []storage.Span{
		makeSpan("t1", "svc", "", 5*time.Millisecond),
		makeSpan("t2", "svc", "", 50*time.Millisecond),
		makeSpan("t3", "svc", "", 200*time.Millisecond),
	}
	c := filter.Criteria{MinDuration: 10 * time.Millisecond, MaxDuration: 100 * time.Millisecond}
	got := filter.Filter(spans, c)
	if len(got) != 1 {
		t.Fatalf("expected 1 span in range, got %d", len(got))
	}
}

func TestFilter_NoCriteria(t *testing.T) {
	spans := []storage.Span{
		makeSpan("t1", "a", "", 1*time.Millisecond),
		makeSpan("t2", "b", "", 2*time.Millisecond),
	}
	got := filter.Filter(spans, filter.Criteria{})
	if len(got) != 2 {
		t.Fatalf("expected all spans returned, got %d", len(got))
	}
}

func TestFilterTraces(t *testing.T) {
	store := storage.NewTraceStore()
	store.AddSpan(makeSpan("trace-1", "auth", "", 10*time.Millisecond))
	store.AddSpan(makeSpan("trace-2", "payment", "err", 30*time.Millisecond))

	got := filter.FilterTraces(store, filter.Criteria{OnlyErrors: true})
	if len(got) != 1 {
		t.Fatalf("expected 1 trace with errors, got %d", len(got))
	}
	if _, ok := got["trace-2"]; !ok {
		t.Fatal("expected trace-2 in results")
	}
}
