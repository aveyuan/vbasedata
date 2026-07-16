package vbasedata

import (
	"time"

	redis "github.com/redis/go-redis/v9"
)

// Cache holds short-lived string values used for verification codes and
// rate-limiting counters. Get and Verify can consume the value atomically.
type Cache interface {
	Set(id, value string) error
	Get(id string, clear bool) string
	Incr(id string) error
	Verify(id, answer string, clear bool) bool
}

// NewCache uses Redis when a client is configured; otherwise it keeps values
// in an in-process LRU cache.
func NewCache(client redis.UniversalClient, size int, expiration time.Duration) Cache {
	if client != nil {
		return NewRedisCache(client, expiration)
	}
	return NewLruCache(size, expiration)
}
