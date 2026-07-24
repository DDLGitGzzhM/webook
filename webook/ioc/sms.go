package ioc

import (
	"time"

	"github.com/redis/go-redis/v9"

	slide "webook/webook/internal/pkg/ratelimit"
	"webook/webook/internal/service/ratelimit"
	"webook/webook/internal/service/sms"
	// "webook/webook/internal/service/sms/metrics"
	smsTest "webook/webook/internal/service/sms/test"
)

func InitSMSService(cmd redis.Cmdable) sms.Service {
	svc := smsTest.NewService()
	// 接入监控
	// return metrics.NewPrometheusDecorator(svc)
	return ratelimit.NewRateLimitSMSService(svc, slide.NewRedisSlideWindowLimiter(cmd, 1*time.Second, 10))
}
