package tracereplay

import (
	"testing"
	"time"

	"github.com/your-org/grpc-tracer/internal/storage"
)

func TestSortByStart_OrdersAscending(t *testing.T) {
	now := time.Now()
	spans := []storage.Span{
		{SpanID: "c", StartTime: now.Add(20 * time.Millisecond)},
		{SpanID: "a", StartTime: now},
		{SpanID: "b", StartTime: now.Add(10 * time.Millisecond)},
	}
	sortByStart(spans)
	expected := []string{"a", "b", "c"}
	for i, sp := range spans {
		if sp.SpanID != expected[i] {
			t.Errorf("pos %d: want %s got %s", i, expected[i], sp.SpanID)
		}
	}
}

func TestSortByStart_SingleElement(t *testing.T) {
	spans := []storage.Span{{SpanID: "only", StartTime: time.Now()}}
	sortByStart(spans) // must not panic
	if spans[0].SpanID != "only" {
		t.Errorf("unexpected mutation")
	}
}

func TestSortByStart_AlreadySorted(t *testing.T) {
	now := time.Now()
	spans := []storage.Span{
		{SpanID: "x", StartTime: now},
		{SpanID: "y", StartTime: now.Add(5 * time.Millisecond)},
	}
	sortByStart(spans)
	if spans[0].SpanID != "x" || spans[1].SpanID != "y" {
		t.Error("already-sorted slice was mutated incorrectly")
	}
}
