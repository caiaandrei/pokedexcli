package pokecache

import (
	"sync"
	"time"
)

type Cache struct {
	entries  map[string]cacheEntry
	interval time.Duration
	sync.Mutex
}

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

func NewCache(interval time.Duration) *Cache {
	cache := &Cache{
		interval: interval,
		entries:  make(map[string]cacheEntry),
	}

	go cache.reapLoop()
	//fmt.Println("Info - Cache created!")
	return cache
}

func (c *Cache) Add(key string, val []byte) {
	c.Lock()
	defer c.Unlock()

	c.entries[key] = cacheEntry{
		createdAt: time.Now(),
		val:       val,
	}

	//fmt.Printf("Info - key updated - %v + %v - count %v\n", key, string(val), len(c.entries))
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.Lock()
	defer c.Unlock()

	entry, ok := c.entries[key]

	if !ok {
		//fmt.Printf("Info - key doesn't exist - %v - count %v\n", key, len(c.entries))
		return nil, false
	}
	//fmt.Printf("Info - key returned - %v + %v\n", key, string(entry.val))
	return entry.val, true
}

func (c *Cache) reapLoop() {
	ticker := time.NewTicker(c.interval)

	for range ticker.C {
		c.reap()
	}
}

func (c *Cache) reap() {
	c.Lock()
	defer c.Unlock()

	now := time.Now()

	for key, entry := range c.entries {
		if now.Sub(entry.createdAt) > c.interval {
			delete(c.entries, key)
			//fmt.Printf("Info - key deleted - %v - count %v\n", key, len(c.entries))
		}
	}
}
