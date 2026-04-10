package tracereplay

import (
	"sort"

	"github.com/your-org/grpc-tracer/internal/storage"
)

// sortByStart sorts a slice of spans in ascending StartTime order.
func sortByStart(spans []storage.Span) {
	sort.Slice(spans, func(i, j int) bool {
		return spans[i].StartTime.Before(spans[j].StartTime)
	})
}
