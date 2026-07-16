package vbasedata

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestRedisCache_LocalEnv(t *testing.T) {
	requireIntegration(t)
	addr, usingDefault := testEnvDefault("VB_TEST_REDIS_ADDR", "127.0.0.1:6379")
	rdb, closeFn, err := NewRedis(&RedisConfig{
		Addr:         []string{addr},
		Auth:         os.Getenv("VB_TEST_REDIS_AUTH"),
		DB:           testEnvIntDefault("VB_TEST_REDIS_DB", 0),
		PoolSize:     2,
		MaxIdle:      1,
		ReadTimeout:  2,
		WriteTimeout: 2,
		MaxIdleTime:  30,
	}, discardLogger())
	if err != nil {
		if usingDefault {
			t.Skipf("redis is not available at %s with default test credentials: %v", addr, err)
		}
		t.Fatalf("NewRedis: %v", err)
	}
	defer closeFn()

	key := fmt.Sprintf("vbasedata:test:cache:%d", time.Now().UnixNano())
	t.Cleanup(func() { _ = rdb.Del(context.Background(), key).Err() })

	cache := NewCache(rdb, 1, time.Minute)
	if _, ok := cache.(*RedisCache); !ok {
		t.Fatalf("NewCache returned %T, want *RedisCache", cache)
	}
	if err := cache.Set(key, "answer"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if !cache.Verify(key, "answer", true) {
		t.Fatal("Verify did not accept stored answer")
	}
	if cache.Verify(key, "answer", false) {
		t.Fatal("Verify accepted an already-consumed answer")
	}

	if err := cache.Set(key, "invalid"); err != nil {
		t.Fatalf("Set invalid value: %v", err)
	}
	if err := cache.Incr(key); err != nil {
		t.Fatalf("Incr invalid value: %v", err)
	}
	if got := cache.Get(key, false); got != "1" {
		t.Fatalf("Incr invalid value = %q, want 1", got)
	}
	if err := cache.Incr(key); err != nil {
		t.Fatalf("second Incr: %v", err)
	}
	if got := cache.Get(key, false); got != "2" {
		t.Fatalf("second Incr = %q, want 2", got)
	}
}

func TestNewCache_FallsBackToLRU(t *testing.T) {
	cache := NewCache(nil, 1, time.Minute)
	if _, ok := cache.(*LruCache); !ok {
		t.Fatalf("NewCache(nil) returned %T, want *LruCache", cache)
	}
}
