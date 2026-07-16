package ratelimit

import (
	"context"
	"fmt"

	"webook/webook/internal/pkg/ratelimit"
	"webook/webook/internal/service/sms"
)

type RateLimitSMSService struct {
	svc   sms.Service
	limit ratelimit.Limiter
}

func NewRateLimitSMSService(svc sms.Service, limit ratelimit.Limiter) *RateLimitSMSService {
	return &RateLimitSMSService{
		svc:   svc,
		limit: limit,
	}
}

func (s RateLimitSMSService) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	limited, err := s.limit.Limited(ctx, "sms:tencent")
	if err != nil {
		return fmt.Errorf("限流错误: %w", err)
	}
	if limited {
		return fmt.Errorf("限流了: %w", err)
	}
	// 增加新特性
	if err = s.svc.Send(ctx, tpl, args, numbers...); err != nil {
		return nil
	}
	// 增加新特性
	return err
}
