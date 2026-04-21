// Package traceforwarder provides a Forwarder that copies spans from a source
// TraceStore to one or more destination TraceStores.
//
// An optional Predicate function can be supplied to gate which spans are
// forwarded; if no predicate is given every span is forwarded unconditionally.
//
// Basic usage:
//
//	f, err := traceforwarder.New(sourceStore, nil, destA, destB)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Forward a single trace
//	n, err := f.ForwardTrace("abc123")
//
//	// Forward all traces in the source
//	n, err = f.ForwardAll()
package traceforwarder
