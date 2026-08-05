package ioc

import (
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

// InitRedis 初始化 Redis
func InitRedis() redis.Cmdable {
	addr := viper.GetString("redis.addr")
	redisClient := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return redisClient
}
