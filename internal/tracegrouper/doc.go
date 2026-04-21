// Package tracegrouper provides utilities for grouping spans from a trace
// store into named buckets based on a configurable key function.
//
// A Grouper reads all spans from a TraceStore and partitions them using a
// KeyFunc — a function that maps a Span to a string key. Two built-in key
// functions are provided:
//
//   - ByService: groups spans by their Service field.
//   - ByMethod:  groups spans by their Method field.
//
// Custom key functions can be supplied to group by any span attribute,
// including tags, trace ID, or derived values.
//
// Example:
//
//	g, err := tracegrouper.New(store, tracegrouper.ByService)
//	if err != nil {
//		log.Fatal(err)
//	}
//	groups, err := g.Group()
//	for service, spans := range groups {
//		fmt.Printf("%s: %d spans\n", service, len(spans))
//	}
package tracegrouper
