package service

import (
	"context"
	"webook/webook/account/domain"
	"webook/webook/account/repository"
)

type accountService struct {
	repo repository.AccountRepository
}

func NewAccountService(repo repository.AccountRepository) AccountService {
	return &accountService{repo: repo}
}

func (a *accountService) Credit(ctx context.Context, cr domain.Credit) error {
	// redis 里面看一下有没有这个 biz + biz_id，有就认为已经处理过了
	// 但是最终肯定是利用唯一索引来兜底的
	return a.repo.AddCredit(ctx, cr)
}
