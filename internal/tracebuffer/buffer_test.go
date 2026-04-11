package tracebuffer_test

import (
	"testing"

	"github.com/example/grpc-tracer/internal/storage"
	"github.com/example/grpc-tracer/internal/tracebuffer"
)

func makeSpan(id string) storage.Span {
	return storage.Span{SpanID: id, TraceID: "trace-1", ServiceName: "svc"}
}

func TestNew_PanicsOnZeroCapacity(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for zero capacity")
		}
	}()
	tracebuffer.New(0)
}

func TestPush_And_Pop_FIFO(t *testing.T) {
	b := tracebuffer.New(4)
	b.Push(makeSpan("a"))
	b.Push(makeSpan("b"))
	b.Push(makeSpan("c"))

	s, err := b.Pop()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.SpanID != "a" {
		t.Errorf("expected 'a', got %q", s.SpanID)
	}
}

func TestPop_EmptyBuffer_ReturnsError(t *testing.T) {
	b := tracebuffer.New(2)
	_, err := b.Pop()
	if err != tracebuffer.ErrEmptyBuffer {
		t.Errorf("expected ErrEmptyBuffer, got %v", err)
	}
}

func TestPush_EvictsOldestWhenFull(t *testing.T) {
	b := tracebuffer.New(3)
	b.Push(makeSpan("a"))
	b.Push(makeSpan("b"))
	b.Push(makeSpan("c"))
	b.Push(makeSpan("d")) // evicts "a"

	if b.Len() != 3 {
		t.Fatalf("expected len 3, got %d", b.Len())
	}

	s, _ := b.Pop()
	if s.SpanID != "b" {
		t.Errorf("expected oldest surviving span 'b', got %q", s.SpanID)
	}
}

func TestLen_TracksCount(t *testing.T) {
	b := tracebuffer.New(10)
	if b.Len() != 0 {
		t.Fatalf("expected 0, got %d", b.Len())
	}
	b.Push(makeSpan("x"))
	b.Push(makeSpan("y"))
	if b.Len() != 2 {
		t.Errorf("expected 2, got %d", b.Len())
	}
}

func TestFlush_ReturnsAllInOrder(t *testing.T) {
	b := tracebuffer.New(5)
	ids := []string{"p", "q", "r"}
	for _, id := range ids {
		b.Push(makeSpan(id))
	}

	spans := b.Flush()
	if len(spans) != 3 {
		t.Fatalf("expected 3 spans, got %d", len(spans))
	}
	for i, id := range ids {
		if spans[i].SpanID != id {
			t.Errorf("index %d: expected %q, got %q", i, id, spans[i].SpanID)
		}
	}
	if b.Len() != 0 {
		t.Errorf("expected buffer empty after flush, got %d", b.Len())
	}
}

func TestFlush_EmptyBuffer_ReturnsEmptySlice(t *testing.T) {
	b := tracebuffer.New(4)
	spans := b.Flush()
	if len(spans) != 0 {
		t.Errorf("expected empty slice, got %d elements", len(spans))
	}
}
