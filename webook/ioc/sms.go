package ioc

import (
	"time"

	"github.com/redis/go-redis/v9"

	slide "webook/webook/internal/pkg/ratelimit"
	"webook/webook/internal/service/ratelimit"
	"webook/webook/internal/service/sms"
	smsTest "webook/webook/internal/service/sms/test"
)

func InitSMSService(cmd redis.Cmdable) sms.Service {
	svc := smsTest.NewService()
	return ratelimit.NewRateLimitSMSService(svc, slide.NewRedisSlideWindowLimiter(cmd, 1*time.Second, 10))
}
