package service

import (
	"context"
	"fmt"
	"math/rand/v2"

	"webook/webook/internal/repository"
	"webook/webook/internal/service/sms"
)

type CodeService struct {
	repo   *repository.CodeRepository
	smsSvc sms.Service
}

func NewCodeService(repo *repository.CodeRepository, smsSvc sms.Service) *CodeService {
	return &CodeService{
		repo:   repo,
		smsSvc: smsSvc,
	}
}

func (svc *CodeService) Send(ctx context.Context, biz string, phone string) error {
	code := svc.generateCode()
	err := svc.repo.Store(ctx, biz, phone, code)
	if err != nil {
		return err
	}
	if err = svc.smsSvc.Send(ctx, "验证码", []string{code}, phone); err != nil {
		// 这个地方怎么办
		// 意味着 redis 有这个验证码，但是你没发成功
		// 我能不能删除这个验证码 。 不能！如果超时，你不知道发没发出去
		return err
	}
	return nil
}

func (svc *CodeService) Verify(ctx context.Context, biz string, phone string, expectCode string) (bool, error) {
	isVerify, err := svc.repo.Verify(ctx, biz, phone, expectCode)
	if err != nil {
		return false, err
	}
	return isVerify, nil
}

func (svc *CodeService) generateCode() string {
	return fmt.Sprintf("%6d", rand.IntN(1000000))
}
