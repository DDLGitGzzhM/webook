package ioc

import (
	rlock "github.com/gotomicro/redis-lock"
	redisv9 "github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

func InitRedis() redisv9.Cmdable {
	type Config struct {
		Addr string `yaml:"addr"`
	}
	var cfg Config
	err := viper.UnmarshalKey("redis", &cfg)
	if err != nil {
		return nil
	}
	return redisv9.NewClient(&redisv9.Options{
		Addr: cfg.Addr,
	})
}

func InitRLockClient(cmd redisv9.Cmdable) *rlock.Client {
	return rlock.NewClient(cmd)
}
