package tracehook_test

import (
	"testing"
	"time"

	"github.com/example/grpc-tracer/internal/storage"
	"github.com/example/grpc-tracer/internal/tracehook"
)

func makeSpan(traceID, spanID, service string) storage.Span {
	return storage.Span{
		TraceID:   traceID,
		SpanID:    spanID,
		Service:   service,
		Method:    "/svc/Method",
		StartTime: time.Now(),
		Duration:  10 * time.Millisecond,
	}
}

func TestNew_ReturnsHook(t *testing.T) {
	h := tracehook.New()
	if h == nil {
		t.Fatal("expected non-nil Hook")
	}
}

func TestRegisterPre_NilFuncReturnsError(t *testing.T) {
	h := tracehook.New()
	if err := h.RegisterPre(nil); err == nil {
		t.Fatal("expected error for nil pre-hook")
	}
}

func TestRegisterPost_NilFuncReturnsError(t *testing.T) {
	h := tracehook.New()
	if err := h.RegisterPost(nil); err == nil {
		t.Fatal("expected error for nil post-hook")
	}
}

func TestRegisterPre_InvokedBeforeStore(t *testing.T) {
	h := tracehook.New()
	store := storage.NewTraceStore()

	var order []string
	_ = h.RegisterPre(func(s storage.Span) { order = append(order, "pre") })
	_ = h.RegisterPost(func(s storage.Span) { order = append(order, "post") })

	span := makeSpan("t1", "s1", "svc")
	_ = h.AddSpan(store, span)

	if len(order) != 2 || order[0] != "pre" || order[1] != "post" {
		t.Fatalf("unexpected hook order: %v", order)
	}
}

func TestRegisterPost_InvokedAfterStore(t *testing.T) {
	h := tracehook.New()
	store := storage.NewTraceStore()

	var seenInPost bool
	_ = h.RegisterPost(func(s storage.Span) {
		spans := store.GetTrace(s.TraceID)
		seenInPost = len(spans) > 0
	})

	span := makeSpan("t2", "s2", "svc")
	_ = h.AddSpan(store, span)

	if !seenInPost {
		t.Fatal("post-hook should see span already in store")
	}
}

func TestMultipleHooks_AllInvoked(t *testing.T) {
	h := tracehook.New()
	store := storage.NewTraceStore()

	counter := 0
	for i := 0; i < 3; i++ {
		_ = h.RegisterPre(func(s storage.Span) { counter++ })
		_ = h.RegisterPost(func(s storage.Span) { counter++ })
	}

	_ = h.AddSpan(store, makeSpan("t3", "s3", "svc"))

	if counter != 6 {
		t.Fatalf("expected 6 hook invocations, got %d", counter)
	}
}

func TestAddSpan_SpanPersistedInStore(t *testing.T) {
	h := tracehook.New()
	store := storage.NewTraceStore()

	span := makeSpan("t4", "s4", "svc")
	_ = h.AddSpan(store, span)

	spans := store.GetTrace("t4")
	if len(spans) != 1 {
		t.Fatalf("expected 1 span in store, got %d", len(spans))
	}
}
