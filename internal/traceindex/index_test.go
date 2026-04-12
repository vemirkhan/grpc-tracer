package traceindex_test

import (
	"testing"
	"time"

	"github.com/example/grpc-tracer/internal/storage"
	"github.com/example/grpc-tracer/internal/traceindex"
)

func makeStore(t *testing.T) *storage.TraceStore {
	t.Helper()
	return storage.NewTraceStore()
}

func addSpan(t *testing.T, store *storage.TraceStore, traceID, spanID, service, method string, tags map[string]string) {
	t.Helper()
	store.AddSpan(storage.Span{
		TraceID:     traceID,
		SpanID:      spanID,
		ServiceName: service,
		Method:      method,
		Tags:        tags,
		StartTime:   time.Now(),
		Duration:    time.Millisecond,
	})
}

func TestBuild_ByService(t *testing.T) {
	store := makeStore(t)
	addSpan(t, store, "trace1", "span1", "auth", "/Auth/Login", nil)
	addSpan(t, store, "trace2", "span2", "orders", "/Orders/Get", nil)
	addSpan(t, store, "trace3", "span3", "auth", "/Auth/Logout", nil)

	idx := traceindex.New()
	idx.Build(store)

	got := idx.ByService("auth")
	if len(got) != 2 {
		t.Fatalf("expected 2 trace IDs for auth, got %d", len(got))
	}
}

func TestBuild_ByMethod(t *testing.T) {
	store := makeStore(t)
	addSpan(t, store, "trace1", "span1", "auth", "/Auth/Login", nil)
	addSpan(t, store, "trace2", "span2", "auth", "/Auth/Login", nil)
	addSpan(t, store, "trace3", "span3", "auth", "/Auth/Logout", nil)

	idx := traceindex.New()
	idx.Build(store)

	got := idx.ByMethod("/Auth/Login")
	if len(got) != 2 {
		t.Fatalf("expected 2 trace IDs for /Auth/Login, got %d", len(got))
	}

	got2 := idx.ByMethod("/Auth/Logout")
	if len(got2) != 1 {
		t.Fatalf("expected 1 trace ID for /Auth/Logout, got %d", len(got2))
	}
}

func TestBuild_ByTag(t *testing.T) {
	store := makeStore(t)
	addSpan(t, store, "trace1", "span1", "svc", "/M", map[string]string{"env": "prod"})
	addSpan(t, store, "trace2", "span2", "svc", "/M", map[string]string{"env": "staging"})
	addSpan(t, store, "trace3", "span3", "svc", "/M", map[string]string{"env": "prod"})

	idx := traceindex.New()
	idx.Build(store)

	got := idx.ByTag("env", "prod")
	if len(got) != 2 {
		t.Fatalf("expected 2 trace IDs for env=prod, got %d", len(got))
	}
}

func TestBuild_EmptyStore(t *testing.T) {
	store := makeStore(t)
	idx := traceindex.New()
	idx.Build(store)

	if got := idx.ByService("any"); got != nil {
		t.Fatalf("expected nil for empty store, got %v", got)
	}
}

func TestBuild_Idempotent(t *testing.T) {
	store := makeStore(t)
	addSpan(t, store, "trace1", "span1", "svc", "/M", nil)

	idx := traceindex.New()
	idx.Build(store)
	idx.Build(store) // rebuild should not duplicate

	got := idx.ByService("svc")
	if len(got) != 1 {
		t.Fatalf("expected 1 after rebuild, got %d", len(got))
	}
}
