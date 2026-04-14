package traceparenthook_test

import (
	"testing"
	"time"

	"github.com/example/grpc-tracer/internal/storage"
	"github.com/example/grpc-tracer/internal/traceparenthook"
)

func makeSpan(traceID, spanID string) storage.Span {
	return storage.Span{
		TraceID:   traceID,
		SpanID:    spanID,
		Service:   "svc",
		Method:    "/pkg.Svc/Method",
		StartTime: time.Now(),
		Duration:  time.Millisecond * 5,
	}
}

func TestNew_ReturnsHook(t *testing.T) {
	h := traceparenthook.New()
	if h == nil {
		t.Fatal("expected non-nil Hook")
	}
}

func TestRegisterPre_InvokedBeforeStore(t *testing.T) {
	h := traceparenthook.New()
	var called []string
	h.RegisterPre(func(s storage.Span) { called = append(called, "pre:"+s.SpanID) })

	span := makeSpan("t1", "s1")
	h.RunPre(span)

	if len(called) != 1 || called[0] != "pre:s1" {
		t.Fatalf("unexpected pre calls: %v", called)
	}
}

func TestRegisterPost_InvokedAfterStore(t *testing.T) {
	h := traceparenthook.New()
	var called []string
	h.RegisterPost(func(s storage.Span) { called = append(called, "post:"+s.SpanID) })

	span := makeSpan("t1", "s2")
	h.RunPost(span)

	if len(called) != 1 || called[0] != "post:s2" {
		t.Fatalf("unexpected post calls: %v", called)
	}
}

func TestMultipleHooks_AllInvoked(t *testing.T) {
	h := traceparenthook.New()
	count := 0
	for i := 0; i < 3; i++ {
		h.RegisterPre(func(s storage.Span) { count++ })
	}
	h.RunPre(makeSpan("t2", "s3"))
	if count != 3 {
		t.Fatalf("expected 3 pre hooks called, got %d", count)
	}
}

func TestWrap_CallsPreAndPost(t *testing.T) {
	store := storage.NewTraceStore()
	h := traceparenthook.New()

	var order []string
	h.RegisterPre(func(s storage.Span) { order = append(order, "pre") })
	h.RegisterPost(func(s storage.Span) { order = append(order, "post") })

	span := makeSpan("t3", "s4")
	if err := h.Wrap(store, span); err != nil {
		t.Fatalf("Wrap returned error: %v", err)
	}

	if len(order) != 2 || order[0] != "pre" || order[1] != "post" {
		t.Fatalf("unexpected hook order: %v", order)
	}

	spans := store.GetTrace("t3")
	if len(spans) != 1 {
		t.Fatalf("expected span stored, got %d", len(spans))
	}
}

func TestWrap_PreCalledEvenIfStoreEmpty(t *testing.T) {
	store := storage.NewTraceStore()
	h := traceparenthook.New()

	preCalled := false
	h.RegisterPre(func(s storage.Span) { preCalled = true })

	_ = h.Wrap(store, makeSpan("t4", "s5"))
	if !preCalled {
		t.Fatal("expected pre hook to be called")
	}
}
