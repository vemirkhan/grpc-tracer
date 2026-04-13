package tracesplitter_test

import (
	"testing"
	"time"

	"github.com/your-org/grpc-tracer/internal/storage"
	"github.com/your-org/grpc-tracer/internal/tracesplitter"
)

func makeSpan(traceID, spanID, service string, tags map[string]string) storage.Span {
	return storage.Span{
		TraceID:     traceID,
		SpanID:      spanID,
		ServiceName: service,
		Method:      "/svc/Method",
		StartTime:   time.Now(),
		Duration:    10 * time.Millisecond,
		Tags:        tags,
	}
}

func newStore(spans ...storage.Span) *storage.TraceStore {
	st := storage.NewTraceStore()
	for _, sp := range spans {
		st.AddSpan(sp)
	}
	return st
}

func TestNew_NilSourceReturnsError(t *testing.T) {
	_, err := tracesplitter.New(nil, tracesplitter.ByService)
	if err == nil {
		t.Fatal("expected error for nil source")
	}
}

func TestNew_NilPredicateReturnsError(t *testing.T) {
	st := storage.NewTraceStore()
	_, err := tracesplitter.New(st, nil)
	if err == nil {
		t.Fatal("expected error for nil predicate")
	}
}

func TestSplit_NonExistentTrace(t *testing.T) {
	st := storage.NewTraceStore()
	spl, _ := tracesplitter.New(st, tracesplitter.ByService)
	_, err := spl.Split("missing-trace")
	if err == nil {
		t.Fatal("expected error for missing trace")
	}
}

func TestSplit_ByService(t *testing.T) {
	spans := []storage.Span{
		makeSpan("t1", "s1", "alpha", nil),
		makeSpan("t1", "s2", "beta", nil),
		makeSpan("t1", "s3", "alpha", nil),
	}
	st := newStore(spans...)
	spl, err := tracesplitter.New(st, tracesplitter.ByService)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	buckets, err := spl.Split("t1")
	if err != nil {
		t.Fatalf("split error: %v", err)
	}
	if len(buckets) != 2 {
		t.Fatalf("expected 2 buckets, got %d", len(buckets))
	}
	alphaSpans, _ := buckets["alpha"].GetTrace("t1")
	if len(alphaSpans) != 2 {
		t.Errorf("expected 2 alpha spans, got %d", len(alphaSpans))
	}
	betaSpans, _ := buckets["beta"].GetTrace("t1")
	if len(betaSpans) != 1 {
		t.Errorf("expected 1 beta span, got %d", len(betaSpans))
	}
}

func TestSplit_ByTag_DropsUntaggedSpans(t *testing.T) {
	spans := []storage.Span{
		makeSpan("t2", "s1", "svc", map[string]string{"region": "us"}),
		makeSpan("t2", "s2", "svc", map[string]string{"region": "eu"}),
		makeSpan("t2", "s3", "svc", nil), // no tag → dropped
	}
	st := newStore(spans...)
	spl, _ := tracesplitter.New(st, tracesplitter.ByTag("region"))
	buckets, err := spl.Split("t2")
	if err != nil {
		t.Fatalf("split error: %v", err)
	}
	if len(buckets) != 2 {
		t.Fatalf("expected 2 buckets, got %d", len(buckets))
	}
	if _, ok := buckets["us"]; !ok {
		t.Error("expected 'us' bucket")
	}
	if _, ok := buckets["eu"]; !ok {
		t.Error("expected 'eu' bucket")
	}
}

func TestSplit_EmptyTrace(t *testing.T) {
	// Add a span then build a store that has the trace key but no spans would
	// pass the predicate (all dropped).
	sp := makeSpan("t3", "s1", "svc", nil)
	st := newStore(sp)
	spl, _ := tracesplitter.New(st, func(_ storage.Span) string { return "" })
	buckets, err := spl.Split("t3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(buckets) != 0 {
		t.Errorf("expected 0 buckets when all spans are dropped, got %d", len(buckets))
	}
}
