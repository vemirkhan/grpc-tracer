package tracewatcher_test

import (
	"sync"
	"testing"
	"time"

	"github.com/example/grpc-tracer/internal/storage"
	"github.com/example/grpc-tracer/internal/tracewatcher"
)

func makeSpan(id, service string) storage.Span {
	return storage.Span{
		SpanID:      id,
		ServiceName: service,
		Method:      "/svc/Method",
		StartTime:   time.Now(),
	}
}

func TestNew_ReturnsWatcher(t *testing.T) {
	w := tracewatcher.New()
	if w == nil {
		t.Fatal("expected non-nil watcher")
	}
}

func TestSubscribe_ReceivesEvent(t *testing.T) {
	w := tracewatcher.New()
	var got tracewatcher.Event
	var wg sync.WaitGroup
	wg.Add(1)
	w.Subscribe(func(e tracewatcher.Event) {
		got = e
		wg.Done()
	})
	span := makeSpan("s1", "auth")
	w.Notify(tracewatcher.Event{Kind: tracewatcher.EventSpanAdded, TraceID: "t1", Span: span})
	wg.Wait()
	if got.TraceID != "t1" {
		t.Errorf("expected traceID t1, got %s", got.TraceID)
	}
	if got.Kind != tracewatcher.EventSpanAdded {
		t.Errorf("unexpected event kind: %s", got.Kind)
	}
}

func TestUnsubscribe_StopsNotifications(t *testing.T) {
	w := tracewatcher.New()
	called := 0
	unsub := w.Subscribe(func(e tracewatcher.Event) { called++ })
	unsub()
	w.Notify(tracewatcher.Event{Kind: tracewatcher.EventSpanAdded, TraceID: "t2"})
	if called != 0 {
		t.Errorf("expected 0 calls after unsubscribe, got %d", called)
	}
}

func TestMultipleSubscribers_AllNotified(t *testing.T) {
	w := tracewatcher.New()
	var mu sync.Mutex
	count := 0
	for i := 0; i < 3; i++ {
		w.Subscribe(func(e tracewatcher.Event) {
			mu.Lock()
			count++
			mu.Unlock()
		})
	}
	w.Notify(tracewatcher.Event{Kind: tracewatcher.EventSpanAdded, TraceID: "t3"})
	time.Sleep(10 * time.Millisecond)
	mu.Lock()
	defer mu.Unlock()
	if count != 3 {
		t.Errorf("expected 3 notifications, got %d", count)
	}
}

func TestWatchStore_AddsSpanAndEmitsEvent(t *testing.T) {
	w := tracewatcher.New()
	store := storage.NewTraceStore()
	var received tracewatcher.Event
	var wg sync.WaitGroup
	wg.Add(1)
	w.Subscribe(func(e tracewatcher.Event) {
		received = e
		wg.Done()
	})
	span := makeSpan("sp1", "order")
	w.WatchStore(store, "trace-abc", span)
	wg.Wait()
	if received.TraceID != "trace-abc" {
		t.Errorf("expected trace-abc, got %s", received.TraceID)
	}
	traces := store.GetAllTraces()
	if len(traces) == 0 {
		t.Error("expected span to be stored")
	}
}
