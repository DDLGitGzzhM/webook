package logger

import (
	"context"

	"go.uber.org/zap"

	"webook/webook/internal/service/sms"
)

type Service struct {
	svc sms.Service
}

func (s Service) Send(ctx context.Context, biz string, args []string, numbers ...string) error {
	zap.L().Debug("发送短信", zap.String("biz", biz))
	err := s.svc.Send(ctx, biz, args, numbers...)
	if err != nil {
		zap.L().Warn("发送短信失败", zap.Error(err))
		return err
	}
	return nil
}
