// Package tracevalidator validates individual spans and batches of spans
// against a configurable set of rules before they enter the storage layer.
//
// Basic usage:
//
//	v := tracevalidator.New(tracevalidator.Options{
//		MaxDuration: 30 * time.Second,
//	})
//
//	if err := v.ValidateSpan(span); err != nil {
//		log.Printf("invalid span: %v", err)
//	}
//
// ValidateAll can be used to validate a slice of spans in one call,
// returning a map of index → error for every failing span.
package tracevalidator
