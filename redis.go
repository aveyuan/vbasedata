package vbasedata

import (
	"context"
	"errors"
	"time"

	"github.com/go-kratos/kratos/v2/log"
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
func NewRedis(c *RedisConfig, logger *log.Helper) (redis.UniversalClient, func(), error) {
	if c == nil {
		return nil, nil, errors.New("redis配置参数不能为空")
	}

	logger.Infof("redis配置%+v", c.Addr)
	//逗号分割，兼容单点和集群两种模式。
	rdb := redis.NewUniversalClient(&redis.UniversalOptions{
		PoolSize:         c.PoolSize, //连接池最大
		MaxIdleConns:     c.MaxIdle,
		Addrs:            c.Addr,
		Password:         c.Auth,
		ReadTimeout:      time.Duration(c.ReadTimeout) * time.Second,
		WriteTimeout:     time.Duration(c.WriteTimeout) * time.Second,
		DB:               c.DB,
		MasterName:       c.MasterName,
		SentinelUsername: c.SentinelUsername,
		SentinelPassword: c.SentinelPassword,
		ConnMaxIdleTime:  time.Duration(c.WriteTimeout) * time.Second,
	})
	pong, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		return nil, nil, err
	}
	logger.Infof("redis ping 情况：%v", pong)
	f := func() {
		logger.Info("Redis 连接池关闭")
		if err := rdb.Close(); err != nil {
			logger.Errorf("Redis 连接池关闭失败 %v", err)
		}
	}
	return rdb, f, nil
}
