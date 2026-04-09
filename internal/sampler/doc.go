// Package sampler provides pluggable trace sampling strategies for grpc-tracer.
//
// Three strategies are available:
//
//	- AlwaysSample  – record every trace (useful for development).
//	- NeverSample   – discard every trace (useful for benchmarks).
//	- ProbabilisticSample – record a random fraction of traces determined
//	  by a rate in the range [0.0, 1.0].
//
// Example usage:
//
//	s := sampler.New(sampler.ProbabilisticSample, 0.25)
//	if s.ShouldSample(traceID) {
//	    // record the trace
//	}
package sampler
