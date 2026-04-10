// Package cache provides a lightweight, thread-safe, TTL-based in-memory
// cache intended for short-lived trace span lookups within grpc-tracer.
//
// Usage:
//
//	c := cache.New(30*time.Second, 5*time.Minute)
//	defer c.Stop()
//
//	c.Set(traceID, spanData)
//	if v, ok := c.Get(traceID); ok {
//		// use v
//	}
//
// The background GC goroutine is started when gcInterval > 0 and must be
// stopped via Stop() when the cache is no longer needed to avoid goroutine
// leaks.
package cache
