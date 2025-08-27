package vbasedata

import (
	"strconv"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
)

type LruCache struct {
	lru *expirable.LRU[string, string]
}

func NewLruCache(len int, exp time.Duration) *LruCache {
	return &LruCache{
		lru: expirable.NewLRU[string, string](len, nil, exp),
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
	v, ok := s.lru.Get(id)
	if ok {
		i, err := strconv.Atoi(v)
		if err != nil {
			s.lru.Add(id, "1")
			return nil
		}
		s.lru.Add(id, strconv.Itoa(i))
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
