package visualizer

import (
	"strings"
	"testing"
	"time"

	"grpc-tracer/internal/storage"
)

func TestBuildTree(t *testing.T) {
	now := time.Now()

	spans := []*storage.Span{
		{
			TraceID:     "trace-1",
			SpanID:      "span-1",
			ServiceName: "root",
			Method:      "/api/root",
			StartTime:   now,
			EndTime:     now.Add(100 * time.Millisecond),
		},
		{
			TraceID:     "trace-1",
			SpanID:      "span-2",
			ParentID:    "span-1",
			ServiceName: "child-1",
			Method:      "/api/child1",
			StartTime:   now.Add(10 * time.Millisecond),
			EndTime:     now.Add(50 * time.Millisecond),
		},
		{
			TraceID:     "trace-1",
			SpanID:      "span-3",
			ParentID:    "span-1",
			ServiceName: "child-2",
			Method:      "/api/child2",
			StartTime:   now.Add(60 * time.Millisecond),
			EndTime:     now.Add(90 * time.Millisecond),
		},
	}

	tree := BuildTree(spans)

	if len(tree) != 1 {
		t.Fatalf("expected 1 root node, got %d", len(tree))
	}

	root := tree[0]
	if root.Span.SpanID != "span-1" {
		t.Errorf("expected root span-1, got %s", root.Span.SpanID)
	}

	if len(root.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(root.Children))
	}
}

func TestBuildTree_MultipleRoots(t *testing.T) {
	now := time.Now()

	spans := []*storage.Span{
		{
			TraceID:     "trace-1",
			SpanID:      "span-1",
			ServiceName: "root-1",
			Method:      "/api/root1",
			StartTime:   now,
			EndTime:     now.Add(50 * time.Millisecond),
		},
		{
			TraceID:     "trace-1",
			SpanID:      "span-2",
			ServiceName: "root-2",
			Method:      "/api/root2",
			StartTime:   now.Add(60 * time.Millisecond),
			EndTime:     now.Add(100 * time.Millisecond),
		},
	}

	tree := BuildTree(spans)

	if len(tree) != 2 {
		t.Fatalf("expected 2 root nodes, got %d", len(tree))
	}
}

func TestFormatTree(t *testing.T) {
	store := storage.NewTraceStore()
	v := NewVisualizer(store)

	traceID := "trace-123"
	now := time.Now()

	store.AddSpan(&storage.Span{
		TraceID:     traceID,
		SpanID:      "span-1",
		ServiceName: "api-gateway",
		Method:      "/api/users",
		StartTime:   now,
		EndTime:     now.Add(100 * time.Millisecond),
	})

	store.AddSpan(&storage.Span{
		TraceID:     traceID,
		SpanID:      "span-2",
		ParentID:    "span-1",
		ServiceName: "user-service",
		Method:      "/internal/getUser",
		StartTime:   now.Add(10 * time.Millisecond),
		EndTime:     now.Add(80 * time.Millisecond),
		Error:       "database error",
	})

	result, err := v.FormatTree(traceID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "api-gateway") {
		t.Error("tree should contain api-gateway")
	}

	if !strings.Contains(result, "user-service") {
		t.Error("tree should contain user-service")
	}

	if !strings.Contains(result, "❌") {
		t.Error("tree should contain error marker")
	}
}
