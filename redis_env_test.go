package vbasedata

import (
	"context"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestNewRedis_LocalEnv(t *testing.T) {
	requireIntegration(t)
	addr, usingDefault := testEnvDefault("VB_TEST_REDIS_ADDR", "127.0.0.1:6379")
	auth := os.Getenv("VB_TEST_REDIS_AUTH")
	db := testEnvIntDefault("VB_TEST_REDIS_DB", 0)

	rdb, closeFn, err := NewRedis(&RedisConfig{
		Addr:         []string{addr},
		Auth:         auth,
		DB:           db,
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

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	key := "vbasedata:test:redis-env"
	if err := rdb.Set(ctx, key, "ok", time.Minute).Err(); err != nil {
		t.Fatalf("redis set: %v", err)
	}
	t.Cleanup(func() {
		_ = rdb.Del(context.Background(), key).Err()
	})

	got, err := rdb.Get(ctx, key).Result()
	if err != nil {
		t.Fatalf("redis get: %v", err)
	}
	if got != "ok" {
		t.Fatalf("redis get = %q, want ok", got)
	}
}

func testEnvIntDefault(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return n
}
