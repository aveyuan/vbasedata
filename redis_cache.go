package vbasedata

import (
	"context"
	"errors"
	"time"

	redis "github.com/redis/go-redis/v9"
)

var errRedisCacheClientNil = errors.New("redis cache client is nil")

// RedisCache stores short-lived values in Redis. Expiration is renewed on Set
// and Incr, matching LruCache's write behavior.
type RedisCache struct {
	client     redis.UniversalClient
	expiration time.Duration
}

var _ Cache = (*RedisCache)(nil)

func NewRedisCache(client redis.UniversalClient, expiration time.Duration) *RedisCache {
	return &RedisCache{client: client, expiration: expiration}
}

func (c *RedisCache) Set(id, value string) error {
	if c == nil || c.client == nil {
		return errRedisCacheClientNil
	}
	return c.client.Set(context.Background(), id, value, c.expiration).Err()
}

func (c *RedisCache) Get(id string, clear bool) string {
	if c == nil || c.client == nil {
		return ""
	}

	var result *redis.StringCmd
	if clear {
		result = c.client.GetDel(context.Background(), id)
	} else {
		result = c.client.Get(context.Background(), id)
	}
	value, err := result.Result()
	if err != nil {
		return ""
	}
	return value
}

func (c *RedisCache) Incr(id string) error {
	if c == nil || c.client == nil {
		return errRedisCacheClientNil
	}
	_, err := redisCacheIncr.Run(context.Background(), c.client, []string{id}, c.expirationMilliseconds()).Result()
	return err
}

func (c *RedisCache) Verify(id, answer string, clear bool) bool {
	if c == nil || c.client == nil {
		return false
	}

	var result *redis.StringCmd
	if clear {
		result = c.client.GetDel(context.Background(), id)
	} else {
		result = c.client.Get(context.Background(), id)
	}
	value, err := result.Result()
	return err == nil && value == answer
}

func (c *RedisCache) expirationMilliseconds() int64 {
	if c.expiration <= 0 {
		return 0
	}
	if c.expiration < time.Millisecond {
		return 1
	}
	return c.expiration.Milliseconds()
}

var redisCacheIncr = redis.NewScript(`
local result = redis.pcall('INCR', KEYS[1])
if type(result) == 'table' and result.err then
  result = 1
  if ARGV[1] == '0' then
    redis.call('SET', KEYS[1], result)
  else
    redis.call('SET', KEYS[1], result, 'PX', ARGV[1])
  end
  return result
end
if ARGV[1] ~= '0' then
  redis.call('PEXPIRE', KEYS[1], ARGV[1])
end
return result
`)
