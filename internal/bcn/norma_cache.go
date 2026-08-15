package bcn

import (
	"container/list"
	"sync"
)

// cacheMax bounds each cache size: distinct entries in a single process
// lifetime. On overflow the LEAST recently used entry is evicted — caches
// are a performance aid, correctness comes from ETag revalidation, so
// eviction is always safe.
const cacheMax = 100

// etagEntry holds one cached value with the ETag of the response it came from.
type etagEntry[T any] struct {
	etag  string
	value T
}

// cacheItem is the list element payload: the key (for map removal on
// eviction) and the cached entry.
type cacheItem[T any] struct {
	key   string
	entry etagEntry[T]
}

// etagCache is an in-memory cache with ETag revalidation, keyed by a
// caller-defined string key (norms: "normID@versionDate"; histories:
// "normID"). Eviction is LRU: every get moves the entry to the front, and
// overflow evicts the back (least recently used). Safe for concurrent use.
type etagCache[T any] struct {
	mu      sync.Mutex
	entries map[string]*list.Element
	order   *list.List // front = most recently used
}

// newEtagCache builds an empty cache.
func newEtagCache[T any]() *etagCache[T] {
	return &etagCache[T]{
		entries: make(map[string]*list.Element),
		order:   list.New(),
	}
}

func (c *etagCache[T]) get(key string) (etagEntry[T], bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	el, ok := c.entries[key]
	if !ok {
		return etagEntry[T]{}, false
	}
	c.order.MoveToFront(el)
	return el.Value.(*cacheItem[T]).entry, true
}

func (c *etagCache[T]) put(key string, e etagEntry[T]) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if el, exists := c.entries[key]; exists {
		el.Value.(*cacheItem[T]).entry = e
		c.order.MoveToFront(el)
		return
	}
	if len(c.entries) >= cacheMax {
		back := c.order.Back()
		delete(c.entries, back.Value.(*cacheItem[T]).key)
		c.order.Remove(back)
	}
	c.entries[key] = c.order.PushFront(&cacheItem[T]{key: key, entry: e})
}
