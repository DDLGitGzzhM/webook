package ioc

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"

	"webook/webook/internal/pkg/redisx"
	"webook/webook/internal/repository/cache"
)

// InitUserCache 配合 PrometheusHook 使用
func InitUserCache(client *redis.Client) cache.UserCache {
	client.AddHook(redisx.NewPrometheusHook(
		prometheus.SummaryOpts{
			Namespace: "geekbang_daming",
			Subsystem: "webook",
			Name:      "redis_cmd",
			Help:      "统计 Redis 命令",
			ConstLabels: map[string]string{
				"biz": "user",
			},
		}))
	panic("你别调用")
}
