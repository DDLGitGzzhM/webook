package ioc

import (
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
