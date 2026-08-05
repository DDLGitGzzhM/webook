package startup

import (
	"context"

	"github.com/redis/go-redis/v9"
)

var redisClient redis.Cmdable

// InitRedis 初始化测试用 Redis
func InitRedis() redis.Cmdable {
	if redisClient == nil {
		redisClient = redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		})
		if err := redisClient.Ping(context.Background()).Err(); err != nil {
			panic(err)
		}
	}
	return redisClient
}
