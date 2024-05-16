package LRU

import "container/list"

type Cache struct {
	maxBytes  int64
	nBytes    int64
	ll        *list.List
	cache     map[string]*list.Element
	onEvicted func(key string, value Value)
}

type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}

func newCache(maxBytes int64, onEvicted func(key string, value Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		nBytes:    0,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		onEvicted: onEvicted,
	}
}

func (c *Cache) Get(key string) (value Value, ok bool) {
	if e, ok := c.cache[key]; ok {
		c.ll.MoveToFront(e)
		kv := e.Value.(*entry)
		return kv.value, true
	}
	return
}

func (c *Cache) removeOldest() {
	e := c.ll.Back()
	if e == nil {
		return
	}
	c.ll.Remove(e)
	kv := e.Value.(*entry)
	delete(c.cache, kv.key)
	c.nBytes -= int64(len(kv.key)) + int64(kv.value.Len())
	if c.onEvicted != nil {
		c.onEvicted(kv.key, kv.value)
	}
}

func (c *Cache) Put(key string, value Value) {
	if e, ok := c.cache[key]; ok {
		c.ll.MoveToFront(e)
		kv := e.Value.(*entry)
		c.nBytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		kv := &entry{
			key:   key,
			value: value,
		}
		e := c.ll.PushFront(kv)
		c.cache[key] = e
		c.nBytes += int64(value.Len()) + int64(len(key))
	}

	for c.maxBytes > 0 && c.maxBytes < c.nBytes {
		c.removeOldest()
	}
}
func (c *Cache) Len() int {
	return c.ll.Len()
}
