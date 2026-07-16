package vbasedata

import (
	"strconv"
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
)

var _ Cache = (*LruCache)(nil)

type LruCache struct {
	lru *expirable.LRU[string, string]
	mu  sync.Mutex // 保护多步缓存操作的原子性
}

func NewLruCache(size int, exp time.Duration) *LruCache {
	return &LruCache{
		lru: expirable.NewLRU[string, string](size, nil, exp),
	}
}

func (c *LruCache) Set(id string, value string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.lru.Add(id, value)
	return nil
}

func (c *LruCache) Get(id string, clear bool) string {
	c.mu.Lock()
	defer c.mu.Unlock()

	value, ok := c.lru.Get(id)
	if !ok {
		return ""
	}
	if clear {
		c.lru.Remove(id)
	}
	return value
}

func (c *LruCache) Incr(id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	value, ok := c.lru.Get(id)
	if !ok {
		c.lru.Add(id, "1")
		return nil
	}

	number, err := strconv.Atoi(value)
	if err != nil {
		c.lru.Add(id, "1")
		return nil
	}
	c.lru.Add(id, strconv.Itoa(number+1))
	return nil
}

func (c *LruCache) Verify(id, answer string, clear bool) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	value, ok := c.lru.Get(id)
	if !ok {
		return false
	}
	if clear {
		c.lru.Remove(id)
	}
	return value == answer
}
