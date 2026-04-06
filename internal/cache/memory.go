package cache

import (
	"sync"
	"time"

	"github.com/ermusthofa/flight-aggregator-service/internal/domain"
)

type CacheItem struct {
	Data      []domain.Flight
	ExpiredAt time.Time
}

type MemoryCache struct {
	mu    sync.RWMutex
	store map[string]CacheItem
	ttl   time.Duration
}

func NewMemoryCache(ttl time.Duration) *MemoryCache {
	return &MemoryCache{
		store: make(map[string]CacheItem),
		ttl:   ttl,
	}
}

func (c *MemoryCache) Get(key string) ([]domain.Flight, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, ok := c.store[key]
	if !ok || time.Now().After(item.ExpiredAt) {
		return nil, false
	}

	return item.Data, true
}

func (c *MemoryCache) Set(key string, data []domain.Flight) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.store[key] = CacheItem{
		Data:      data,
		ExpiredAt: time.Now().Add(c.ttl),
	}
}
