// Package tracetopology derives a service dependency graph from the spans
// stored in a TraceStore.
//
// # Overview
//
// A Topology is constructed from a *storage.TraceStore. Calling Build walks
// every recorded span and creates a directed edge from the parent span's
// service to the child span's service whenever they differ. The resulting
// Graph can be queried for edges and service names.
//
// # Example
//
//	topo, err := tracetopology.New(store)
//	if err != nil {
//		log.Fatal(err)
//	}
//	graph := topo.Build()
//	for _, e := range graph.Edges() {
//		fmt.Printf("%s -> %s (%d calls)\n", e.From, e.To, e.Calls)
//	}
package tracetopology
