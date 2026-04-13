// Package tracecache provides a read-through, TTL-based in-memory cache
// for trace lookups backed by a storage.TraceStore.
//
// Usage:
//
//	store := storage.NewTraceStore()
//	cache, err := tracecache.New(store, tracecache.Options{
//		TTL:        10 * time.Second,
//		MaxEntries: 512,
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	spans, ok := cache.GetTrace("my-trace-id")
//
// Entries are evicted when they exceed their TTL or when the cache
// reaches MaxEntries (oldest-first eviction).
package tracecache
