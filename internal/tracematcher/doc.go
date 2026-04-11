// Package tracematcher provides flexible pattern-based matching for gRPC traces
// and spans stored in a TraceStore.
//
// Use [Criteria] to specify one or more conditions — such as service name,
// method, minimum duration, error presence, or tag key/value pairs — then
// pass the criteria to [Match] for a single trace or [MatchAll] to scan an
// entire store.
//
// Criteria fields are combined with AND semantics: a trace must satisfy all
// non-zero fields to be considered a match. String fields such as ServiceName
// and Method are matched case-insensitively as substring matches, so
// ServiceName: "order" will match both "order-service" and "OrderProcessor".
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
