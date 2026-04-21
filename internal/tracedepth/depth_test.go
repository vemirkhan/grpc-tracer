package tracedepth_test

import (
	"testing"
	"time"

	"github.com/example/grpc-tracer/internal/storage"
	"github.com/example/grpc-tracer/internal/tracedepth"
)

func makeSpan(spanID, parentID string) storage.Span {
	return storage.Span{
		TraceID:  "trace-1",
		SpanID:   spanID,
		ParentID: parentID,
		Service:  "svc",
		Method:   "/Test",
		Start:    time.Now(),
		Duration: time.Millisecond,
	}
}

func TestDepth_EmptySpans(t *testing.T) {
	a := tracedepth.New()
	if got := a.Depth(nil); got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}

func TestDepth_SingleRoot(t *testing.T) {
	a := tracedepth.New()
	spans := []storage.Span{makeSpan("s1", "")}
	if got := a.Depth(spans); got != 1 {
		t.Fatalf("expected 1, got %d", got)
	}
}

func TestDepth_LinearChain(t *testing.T) {
	a := tracedepth.New()
	spans := []storage.Span{
		makeSpan("s1", ""),
		makeSpan("s2", "s1"),
		makeSpan("s3", "s2"),
	}
	if got := a.Depth(spans); got != 3 {
		t.Fatalf("expected 3, got %d", got)
	}
}

func TestDepth_BranchingTree(t *testing.T) {
	a := tracedepth.New()
	// root -> s2, s3; s2 -> s4
	spans := []storage.Span{
		makeSpan("root", ""),
		makeSpan("s2", "root"),
		makeSpan("s3", "root"),
		makeSpan("s4", "s2"),
	}
	if got := a.Depth(spans); got != 3 {
		t.Fatalf("expected 3, got %d", got)
	}
}

func TestValidate_NoMaxDepth_AlwaysPasses(t *testing.T) {
	a := tracedepth.New() // MaxDepth == 0 means unlimited
	spans := []storage.Span{
		makeSpan("s1", ""),
		makeSpan("s2", "s1"),
		makeSpan("s3", "s2"),
		makeSpan("s4", "s3"),
		makeSpan("s5", "s4"),
	}
	if err := a.Validate(spans); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestValidate_WithinLimit(t *testing.T) {
	a := tracedepth.New(tracedepth.WithMaxDepth(5))
	spans := []storage.Span{
		makeSpan("s1", ""),
		makeSpan("s2", "s1"),
	}
	if err := a.Validate(spans); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestValidate_ExceedsLimit(t *testing.T) {
	a := tracedepth.New(tracedepth.WithMaxDepth(2))
	spans := []storage.Span{
		makeSpan("s1", ""),
		makeSpan("s2", "s1"),
		makeSpan("s3", "s2"),
	}
	if err := a.Validate(spans); err == nil {
		t.Fatal("expected ErrDepthExceeded, got nil")
	}
}
