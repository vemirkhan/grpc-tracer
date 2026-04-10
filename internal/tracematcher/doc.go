// Package tracematcher provides flexible pattern-based matching for gRPC traces
// and spans stored in a TraceStore.
//
// Use [Criteria] to specify one or more conditions — such as service name,
// method, minimum duration, error presence, or tag key/value pairs — then
// pass the criteria to [Match] for a single trace or [MatchAll] to scan an
// entire store.
//
// Example:
//
//	c := tracematcher.Criteria{
//		ServiceName: "order",
//		OnlyErrors:  true,
//		MinDuration: 200 * time.Millisecond,
//	}
//	matches := tracematcher.MatchAll(store, c)
package tracematcher
