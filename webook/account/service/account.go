package service

import (
	"context"

	"webook/webook/account/domain"
	"webook/webook/account/repository"
)

type accountService struct {
	repo repository.AccountRepository
}

// NewAccountService 创建账户服务
func NewAccountService(repo repository.AccountRepository) AccountService {
	return &accountService{repo: repo}
}

func (a *accountService) Credit(ctx context.Context, cr domain.Credit) error {
	err := a.repo.CheckUnique(ctx, cr)
	if err != nil {
		return err
	}
	// 我这里是有唯一索引的
	err = a.repo.AddCredit(ctx, cr)
	if err == nil {
		// 注意这些部分失败是没有什么问题的
		// 因为我们始终有一个兜底，就是唯一索引。
		err1 := a.repo.SetUnique(ctx, cr)
		if err1 != nil {
			// 记录日志
		}
	}
	return err
}
