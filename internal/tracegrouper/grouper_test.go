package tracegrouper_test

import (
	"testing"
	"time"

	"github.com/user/grpc-tracer/internal/storage"
	"github.com/user/grpc-tracer/internal/tracegrouper"
)

func makeStore() *storage.TraceStore {
	return storage.NewTraceStore()
}

func addSpan(s *storage.TraceStore, traceID, spanID, service, method string, dur time.Duration) {
	s.AddSpan(storage.Span{
		TraceID:   traceID,
		SpanID:    spanID,
		Service:   service,
		Method:    method,
		StartTime: time.Now(),
		Duration:  dur,
	})
}

func TestNew_NilStoreReturnsError(t *testing.T) {
	_, err := tracegrouper.New(nil, tracegrouper.ByService)
	if err == nil {
		t.Fatal("expected error for nil store")
	}
}

func TestNew_NilKeyFuncReturnsError(t *testing.T) {
	s := makeStore()
	_, err := tracegrouper.New(s, nil)
	if err == nil {
		t.Fatal("expected error for nil key func")
	}
}

func TestNew_ValidGrouper(t *testing.T) {
	s := makeStore()
	g, err := tracegrouper.New(s, tracegrouper.ByService)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g == nil {
		t.Fatal("expected non-nil grouper")
	}
}

func TestGroup_ByService(t *testing.T) {
	s := makeStore()
	addSpan(s, "t1", "s1", "auth", "/Login", 10*time.Millisecond)
	addSpan(s, "t2", "s2", "auth", "/Logout", 5*time.Millisecond)
	addSpan(s, "t3", "s3", "billing", "/Pay", 20*time.Millisecond)

	g, _ := tracegrouper.New(s, tracegrouper.ByService)
	groups, err := g.Group()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups["auth"]) != 2 {
		t.Errorf("expected 2 auth spans, got %d", len(groups["auth"]))
	}
	if len(groups["billing"]) != 1 {
		t.Errorf("expected 1 billing span, got %d", len(groups["billing"]))
	}
}

func TestGroup_ByMethod(t *testing.T) {
	s := makeStore()
	addSpan(s, "t1", "s1", "auth", "/Login", 10*time.Millisecond)
	addSpan(s, "t2", "s2", "auth", "/Login", 8*time.Millisecond)
	addSpan(s, "t3", "s3", "billing", "/Pay", 20*time.Millisecond)

	g, _ := tracegrouper.New(s, tracegrouper.ByMethod)
	groups, err := g.Group()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups["/Login"]) != 2 {
		t.Errorf("expected 2 /Login spans, got %d", len(groups["/Login"]))
	}
	if len(groups["/Pay"]) != 1 {
		t.Errorf("expected 1 /Pay span, got %d", len(groups["/Pay"]))
	}
}

func TestGroup_EmptyStore(t *testing.T) {
	s := makeStore()
	g, _ := tracegrouper.New(s, tracegrouper.ByService)
	groups, err := g.Group()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 0 {
		t.Errorf("expected empty groups, got %d", len(groups))
	}
}
