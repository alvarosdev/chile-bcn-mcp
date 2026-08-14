package bcn

import "sync"

// normaCacheMax bounds the cache size: distinct norms requested in a single
// process lifetime. On overflow an arbitrary entry is evicted (Go map
// iteration) — the cache is a performance aid, correctness comes from ETag
// revalidation, so eviction is always safe.
const normaCacheMax = 100

// cacheEntry holds one cached norm with the ETag of the response it came from.
type cacheEntry struct {
	etag  string
	norma NormaFull
}

// NormaCache is an in-memory cache for norms keyed by norm id, with ETag
// revalidation. Safe for concurrent use.
type NormaCache struct {
	mu      sync.Mutex
	entries map[int64]cacheEntry
}

// NewNormaCache builds an empty cache.
func NewNormaCache() *NormaCache {
	return &NormaCache{entries: make(map[int64]cacheEntry)}
}

func (c *NormaCache) get(normID int64) (cacheEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[normID]
	return e, ok
}

func (c *NormaCache) put(normID int64, e cacheEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, exists := c.entries[normID]; !exists && len(c.entries) >= normaCacheMax {
		for k := range c.entries { // arbitrary eviction
			delete(c.entries, k)
			break
		}
	}
	c.entries[normID] = e
}
