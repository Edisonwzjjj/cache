package geecache

import (
	"cache/geecache/LRU"
	"sync"
)

type Cache struct {
	mu         sync.RWMutex
	lru        *LRU.LruCache
	cacheBytes int64
}

func (c *Cache) Put(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		c.lru = LRU.NewCache(c.cacheBytes, nil)
	}
	c.lru.Put(key, value)
}

func (c *Cache) Get(key string) (b ByteView, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.lru == nil {
		return ByteView{}, false
	}
	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}
	return
}
