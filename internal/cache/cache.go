package cache

import (
	"container/list"
	"time"
)

type Item struct {
	Key       string
	Value     interface{}
	ExpiresAt time.Time
}

type Cache struct {
	Capacity  int
	TTL       int
	cacheMap  map[string]*list.Element
	cacheList *list.List
}

func NewCache(capacity int, ttlMillis int) *Cache {
	return &Cache{
		Capacity:  capacity,
		TTL:       ttlMillis,
		cacheMap:  make(map[string]*list.Element),
		cacheList: list.New(),
	}
}

func (c *Cache) Get(key string) (interface{}, bool) {
	if elem, found := c.cacheMap[key]; found {
		cacheItem := elem.Value.(*Item)
		if time.Now().Before(cacheItem.ExpiresAt) {
			c.cacheList.MoveToFront(elem)
			return cacheItem.Value, true
		}
		c.removeElement(elem)
	}
	return nil, false
}

func (c *Cache) Set(key string, value interface{}) {
	expiresAt := time.Now().Add(time.Millisecond * time.Duration(c.TTL))
	if elem, found := c.cacheMap[key]; found {
		c.cacheList.MoveToFront(elem)
		cacheItem := elem.Value.(*Item)
		cacheItem.Value = value
		cacheItem.ExpiresAt = expiresAt
	} else {
		if c.cacheList.Len() >= c.Capacity {
			// Evict the least recently used item
			backElem := c.cacheList.Back()
			if backElem != nil {
				c.removeElement(backElem)
			}
		}
		cacheItem := &Item{
			Key:       key,
			Value:     value,
			ExpiresAt: expiresAt,
		}
		newElem := c.cacheList.PushFront(cacheItem)
		c.cacheMap[key] = newElem
	}
}

func (c *Cache) removeElement(elem *list.Element) {
	c.cacheList.Remove(elem)
	cacheItem := elem.Value.(*Item)
	delete(c.cacheMap, cacheItem.Key)
}
