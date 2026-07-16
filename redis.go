package vbasedata

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	redis "github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Addr             []string `json:"addr" yaml:"addr"`                           // redis地址
	Auth             string   `json:"auth" yaml:"auth"`                           // redis密码
	PoolSize         int      `json:"pool_size" yaml:"pool_size"`                 //连接池最大
	MaxIdle          int      `json:"max_idle" yaml:"max_idle"`                   //空闲连接数
	ReadTimeout      int      `json:"read_timeout" yaml:"read_timeout"`           // 读取超时时间，单位秒
	WriteTimeout     int      `json:"write_timeout" yaml:"write_timeout"`         // 写入超时时间，单位秒
	MaxIdleTime      int      `json:"max_idle_time" yaml:"max_idle_time"`         // 最大空闲时间
	DB               int      `json:"db" yaml:"db"`                               // redis数据库
	MasterName       string   `json:"master_name" yaml:"master_name"`             //哨兵模式下的主节点名称
	SentinelUsername string   `json:"sentinel_username" yaml:"sentinel_username"` //哨兵模式下的用户名
	SentinelPassword string   `json:"sentinel_password" yaml:"sentinel_password"` //哨兵模式下的密码
}

// NewRedis redis连接
func NewRedis(c *RedisConfig, logger *slog.Logger) (redis.UniversalClient, func(), error) {
	if c == nil {
		return nil, nil, errors.New("redis配置参数不能为空")
	}
	if len(c.Addr) == 0 {
		return nil, nil, errors.New("redis地址不能为空")
	}
	for _, addr := range c.Addr {
		if strings.TrimSpace(addr) == "" {
			return nil, nil, errors.New("redis地址不能为空")
		}
	}
	if logger == nil {
		logger = slog.Default()
	}

	logger.Info("初始化 Redis 客户端", "addresses", c.Addr)
	rdb := redis.NewUniversalClient(&redis.UniversalOptions{
		PoolSize:         c.PoolSize,
		MaxIdleConns:     c.MaxIdle,
		Addrs:            c.Addr,
		Password:         c.Auth,
		ReadTimeout:      time.Duration(c.ReadTimeout) * time.Second,
		WriteTimeout:     time.Duration(c.WriteTimeout) * time.Second,
		DB:               c.DB,
		MasterName:       c.MasterName,
		SentinelUsername: c.SentinelUsername,
		SentinelPassword: c.SentinelPassword,
		ConnMaxIdleTime:  time.Duration(c.MaxIdleTime) * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		_ = rdb.Close()
		return nil, nil, fmt.Errorf("ping redis: %w", err)
	}
	logger.Info("Redis ping 成功", "result", pong)

	closeFn := func() {
		logger.Info("Redis 连接池关闭")
		if err := rdb.Close(); err != nil {
			logger.Error("Redis 连接池关闭失败", "error", err)
		}
	}
	return rdb, closeFn, nil
}
