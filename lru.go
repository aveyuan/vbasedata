package vbasedata

import (
	"strconv"
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
)

type LruCache struct {
	lru *expirable.LRU[string, string]
	mu  sync.Mutex // 保护 Incr 的读-改-写复合操作
}

func NewLruCache(size int, exp time.Duration) *LruCache {
	return &LruCache{
		lru: expirable.NewLRU[string, string](size, nil, exp),
	}
}

func (s *LruCache) Set(id string, value string) error {
	s.lru.Add(id, value)
	return nil
}

func (s *LruCache) Get(id string, clear bool) string {
	v, ok := s.lru.Get(id)
	if ok {
		if clear {
			s.lru.Remove(id)
		}
		return v
	}
	return ""
}

func (s *LruCache) Incr(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.lru.Get(id)
	if ok {
		i, err := strconv.Atoi(v)
		if err != nil {
			s.lru.Add(id, "1")
			return nil
		}
		s.lru.Add(id, strconv.Itoa(i+1))
		return nil
	}
	s.lru.Add(id, "1")
	return nil
}

func (s *LruCache) Verify(id, answer string, clear bool) bool {
	v, ok := s.lru.Get(id)
	if ok {
		if clear {
			s.lru.Remove(id)
		}
		return v == answer
	}
	return false
}
