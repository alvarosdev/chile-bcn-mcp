package bcn

import "sync"

// cacheMax bounds each cache size: distinct entries in a single process
// lifetime. On overflow an arbitrary entry is evicted (Go map iteration) —
// caches are a performance aid, correctness comes from ETag revalidation,
// so eviction is always safe.
const cacheMax = 100

// etagEntry holds one cached value with the ETag of the response it came from.
type etagEntry[T any] struct {
	etag  string
	value T
}

// etagCache is an in-memory cache with ETag revalidation, keyed by a
// caller-defined string key (norms: "normID@versionDate"; histories:
// "normID"). Safe for concurrent use.
type etagCache[T any] struct {
	mu      sync.Mutex
	entries map[string]etagEntry[T]
}

// newEtagCache builds an empty cache.
func newEtagCache[T any]() *etagCache[T] {
	return &etagCache[T]{entries: make(map[string]etagEntry[T])}
}

func (c *etagCache[T]) get(key string) (etagEntry[T], bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[key]
	return e, ok
}

func (c *etagCache[T]) put(key string, e etagEntry[T]) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, exists := c.entries[key]; !exists && len(c.entries) >= cacheMax {
		for k := range c.entries { // arbitrary eviction
			delete(c.entries, k)
			break
		}
	}
	c.entries[key] = e
}
