// Package tracebuffer provides a fixed-capacity ring buffer for spans.
// When the buffer is full, the oldest span is evicted to make room for
// the newest, ensuring bounded memory usage during high-throughput tracing.
package tracebuffer

import (
	"errors"
	"sync"

	"github.com/example/grpc-tracer/internal/storage"
)

// ErrEmptyBuffer is returned when a Pop is attempted on an empty buffer.
var ErrEmptyBuffer = errors.New("tracebuffer: buffer is empty")

// Buffer is a thread-safe, fixed-capacity ring buffer of spans.
type Buffer struct {
	mu       sync.Mutex
	slots    []storage.Span
	head     int // index of the next write position
	tail     int // index of the next read position
	count    int
	capacity int
}

// New creates a Buffer with the given capacity.
// Panics if capacity is less than 1.
func New(capacity int) *Buffer {
	if capacity < 1 {
		panic("tracebuffer: capacity must be at least 1")
	}
	return &Buffer{
		slots:    make([]storage.Span, capacity),
		capacity: capacity,
	}
}

// Push adds a span to the buffer. If the buffer is full, the oldest span
// is silently evicted.
func (b *Buffer) Push(s storage.Span) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.count == b.capacity {
		// evict oldest
		b.tail = (b.tail + 1) % b.capacity
		b.count--
	}
	b.slots[b.head] = s
	b.head = (b.head + 1) % b.capacity
	b.count++
}

// Pop removes and returns the oldest span.
// Returns ErrEmptyBuffer if the buffer contains no spans.
func (b *Buffer) Pop() (storage.Span, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.count == 0 {
		return storage.Span{}, ErrEmptyBuffer
	}
	s := b.slots[b.tail]
	b.tail = (b.tail + 1) % b.capacity
	b.count--
	return s, nil
}

// Len returns the number of spans currently held in the buffer.
func (b *Buffer) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.count
}

// Flush drains all spans from the buffer and returns them in FIFO order.
func (b *Buffer) Flush() []storage.Span {
	b.mu.Lock()
	defer b.mu.Unlock()

	out := make([]storage.Span, b.count)
	for i := 0; i < len(out); i++ {
		out[i] = b.slots[(b.tail+i)%b.capacity]
	}
	b.head = 0
	b.tail = 0
	b.count = 0
	return out
}
