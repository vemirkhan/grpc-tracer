// Package tracecloner implements deep-copy cloning of traces between
// TraceStore instances.
//
// # Overview
//
// A Cloner is constructed with a source and a destination TraceStore.
// It can clone a single trace by ID or bulk-clone every trace present
// in the source:
//
//	c, err := tracecloner.New(src, dest)
//	if err != nil { ... }
//
//	// Clone one trace
//	if err := c.CloneTrace("abc123"); err != nil { ... }
//
//	// Clone everything
//	n := c.CloneAll()
//
// All span tags are deep-copied so that mutations in the source store
// do not affect the cloned spans and vice-versa.
package tracecloner
