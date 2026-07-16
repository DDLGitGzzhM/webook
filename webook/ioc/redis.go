package ioc

import (
	redisv9 "github.com/redis/go-redis/v9"

	"webook/webook/config/config"
)

func InitRedis() redisv9.Cmdable {
	return redisv9.NewClient(&redisv9.Options{
		Addr: config.Config.Redis.Addr,
	})
}
