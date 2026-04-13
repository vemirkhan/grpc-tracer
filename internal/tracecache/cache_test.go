package tracecache_test

import (
	"testing"
	"time"

	"github.com/user/grpc-tracer/internal/storage"
	"github.com/user/grpc-tracer/internal/tracecache"
)

func makeStore(t *testing.T) *storage.TraceStore {
	t.Helper()
	return storage.NewTraceStore()
}

func addSpan(t *testing.T, s *storage.TraceStore, traceID, spanID string) {
	t.Helper()
	s.AddSpan(storage.Span{TraceID: traceID, SpanID: spanID, Service: "svc"})
}

func TestNew_NilStoreReturnsError(t *testing.T) {
	_, err := tracecache.New(nil)
	if err == nil {
		t.Fatal("expected error for nil store")
	}
}

func TestNew_ValidStore(t *testing.T) {
	s := makeStore(t)
	c, err := tracecache.New(s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil cache")
	}
}

func TestGetTrace_Miss(t *testing.T) {
	s := makeStore(t)
	c, _ := tracecache.New(s)
	_, ok := c.GetTrace("nonexistent")
	if ok {
		t.Fatal("expected miss for unknown traceID")
	}
}

func TestGetTrace_Hit(t *testing.T) {
	s := makeStore(t)
	addSpan(t, s, "trace-1", "span-1")
	c, _ := tracecache.New(s)

	spans, ok := c.GetTrace("trace-1")
	if !ok {
		t.Fatal("expected hit")
	}
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if c.Len() != 1 {
		t.Fatalf("expected cache len 1, got %d", c.Len())
	}
}

func TestGetTrace_CachedSecondCall(t *testing.T) {
	s := makeStore(t)
	addSpan(t, s, "trace-2", "span-2")
	c, _ := tracecache.New(s)

	c.GetTrace("trace-2")
	// Second call should still succeed (from cache)
	spans, ok := c.GetTrace("trace-2")
	if !ok || len(spans) != 1 {
		t.Fatal("expected cached result on second call")
	}
}

func TestGetTrace_TTLExpiry(t *testing.T) {
	s := makeStore(t)
	addSpan(t, s, "trace-3", "span-3")
	c, _ := tracecache.New(s, tracecache.Options{TTL: 1 * time.Millisecond})

	c.GetTrace("trace-3")
	time.Sleep(5 * time.Millisecond)
	// After TTL, cache should re-fetch
	_, ok := c.GetTrace("trace-3")
	if !ok {
		t.Fatal("expected re-fetch after TTL expiry")
	}
}

func TestInvalidate_RemovesEntry(t *testing.T) {
	s := makeStore(t)
	addSpan(t, s, "trace-4", "span-4")
	c, _ := tracecache.New(s)

	c.GetTrace("trace-4")
	if c.Len() != 1 {
		t.Fatal("expected 1 cached entry")
	}
	c.Invalidate("trace-4")
	if c.Len() != 0 {
		t.Fatal("expected 0 cached entries after invalidation")
	}
}

func TestMaxEntries_EvictsOldest(t *testing.T) {
	s := makeStore(t)
	for i := 0; i < 3; i++ {
		addSpan(t, s, string(rune('a'+i)), "span")
	}
	c, _ := tracecache.New(s, tracecache.Options{MaxEntries: 2})

	c.GetTrace("a")
	c.GetTrace("b")
	c.GetTrace("c") // should evict "a"

	if c.Len() != 2 {
		t.Fatalf("expected 2 entries, got %d", c.Len())
	}
}
