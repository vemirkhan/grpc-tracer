package tracepartitioner_test

import (
	"testing"
	"time"

	"github.com/example/grpc-tracer/internal/storage"
	"github.com/example/grpc-tracer/internal/tracepartitioner"
)

func makeStore() *storage.TraceStore { return storage.NewTraceStore() }

func addSpan(s *storage.TraceStore, traceID, spanID, service, method string) {
	s.AddSpan(storage.Span{
		TraceID:   traceID,
		SpanID:    spanID,
		Service:   service,
		Method:    method,
		StartTime: time.Now(),
		Duration:  time.Millisecond,
	})
}

func TestNew_NilSourceReturnsError(t *testing.T) {
	_, err := tracepartitioner.New(nil, tracepartitioner.ByService)
	if err == nil {
		t.Fatal("expected error for nil source")
	}
}

func TestNew_NilKeyFuncReturnsError(t *testing.T) {
	_, err := tracepartitioner.New(makeStore(), nil)
	if err == nil {
		t.Fatal("expected error for nil key function")
	}
}

func TestNew_ValidPartitioner(t *testing.T) {
	p, err := tracepartitioner.New(makeStore(), tracepartitioner.ByService)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p == nil {
		t.Fatal("expected non-nil partitioner")
	}
}

func TestPartition_ByService(t *testing.T) {
	src := makeStore()
	addSpan(src, "t1", "s1", "auth", "/Login")
	addSpan(src, "t1", "s2", "auth", "/Logout")
	addSpan(src, "t2", "s3", "billing", "/Charge")

	p, _ := tracepartitioner.New(src, tracepartitioner.ByService)
	parts := p.Partition()

	if len(parts) != 2 {
		t.Fatalf("expected 2 partitions, got %d", len(parts))
	}
	authSpans := parts["auth"].GetTrace("t1")
	if len(authSpans) != 2 {
		t.Errorf("expected 2 auth spans, got %d", len(authSpans))
	}
	billingSpans := parts["billing"].GetTrace("t2")
	if len(billingSpans) != 1 {
		t.Errorf("expected 1 billing span, got %d", len(billingSpans))
	}
}

func TestPartition_ByMethod(t *testing.T) {
	src := makeStore()
	addSpan(src, "t1", "s1", "svc", "/Ping")
	addSpan(src, "t2", "s2", "svc", "/Pong")
	addSpan(src, "t3", "s3", "svc", "/Ping")

	p, _ := tracepartitioner.New(src, tracepartitioner.ByMethod)
	parts := p.Partition()

	if len(parts) != 2 {
		t.Fatalf("expected 2 partitions, got %d", len(parts))
	}
	if len(parts["/Ping"].GetAllTraces()) != 2 {
		t.Errorf("expected 2 traces under /Ping")
	}
}

func TestPartition_EmptyKeyFallsToDefault(t *testing.T) {
	src := makeStore()
	addSpan(src, "t1", "s1", "", "")

	emptyKey := func(s storage.Span) string { return "" }
	p, _ := tracepartitioner.New(src, emptyKey)
	parts := p.Partition()

	if _, ok := parts["__default__"]; !ok {
		t.Fatal("expected __default__ partition for empty key")
	}
}

func TestKeys_ReflectsLastPartition(t *testing.T) {
	src := makeStore()
	addSpan(src, "t1", "s1", "x", "/A")
	addSpan(src, "t2", "s2", "y", "/B")

	p, _ := tracepartitioner.New(src, tracepartitioner.ByService)
	p.Partition()
	keys := p.Keys()

	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}
}
