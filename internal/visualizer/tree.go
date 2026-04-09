package visualizer

import (
	"fmt"
	"sort"
	"strings"

	"grpc-tracer/internal/storage"
)

// TreeNode represents a node in the trace tree
type TreeNode struct {
	Span     *storage.Span
	Children []*TreeNode
}

// BuildTree constructs a tree structure from spans
func BuildTree(spans []*storage.Span) []*TreeNode {
	if len(spans) == 0 {
		return nil
	}

	// Create a map of spanID to node
	nodeMap := make(map[string]*TreeNode)
	for _, span := range spans {
		nodeMap[span.SpanID] = &TreeNode{
			Span:     span,
			Children: make([]*TreeNode, 0),
		}
	}

	// Build parent-child relationships
	roots := make([]*TreeNode, 0)
	for _, node := range nodeMap {
		if node.Span.ParentID == "" {
			roots = append(roots, node)
		} else if parent, exists := nodeMap[node.Span.ParentID]; exists {
			parent.Children = append(parent.Children, node)
		} else {
			// Parent not found, treat as root
			roots = append(roots, node)
		}
	}

	// Sort roots by start time
	sort.Slice(roots, func(i, j int) bool {
		return roots[i].Span.StartTime.Before(roots[j].Span.StartTime)
	})

	return roots
}

// FormatTree returns a tree-style visualization of the trace
func (v *Visualizer) FormatTree(traceID string) (string, error) {
	spans, err := v.store.GetTrace(traceID)
	if err != nil {
		return "", err
	}

	if len(spans) == 0 {
		return "", fmt.Errorf("no spans found for trace %s", traceID)
	}

	tree := BuildTree(spans)
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("\n=== Trace Tree: %s ===\n\n", traceID))

	for _, root := range tree {
		formatNode(&builder, root, "", true)
	}

	return builder.String(), nil
}

func formatNode(builder *strings.Builder, node *TreeNode, prefix string, isLast bool) {
	duration := node.Span.EndTime.Sub(node.Span.StartTime)
	
	// Draw tree structure
	if prefix == "" {
		builder.WriteString("└── ")
	} else {
		if isLast {
			builder.WriteString(prefix + "└── ")
		} else {
			builder.WriteString(prefix + "├── ")
		}
	}

	// Format span info
	errorMarker := ""
	if node.Span.Error != "" {
		errorMarker = " ❌"
	}
	
	builder.WriteString(fmt.Sprintf("%s [%s] %s (%v)%s\n",
		node.Span.ServiceName,
		node.Span.SpanID[:8],
		node.Span.Method,
		duration,
		errorMarker,
	))

	// Sort children by start time
	sort.Slice(node.Children, func(i, j int) bool {
		return node.Children[i].Span.StartTime.Before(node.Children[j].Span.StartTime)
	})

	// Recursively format children
	for i, child := range node.Children {
		newPrefix := prefix
		if prefix == "" {
			newPrefix = "    "
		} else {
			if isLast {
				newPrefix = prefix + "    "
			} else {
				newPrefix = prefix + "│   "
			}
		}
		formatNode(builder, child, newPrefix, i == len(node.Children)-1)
	}
}
