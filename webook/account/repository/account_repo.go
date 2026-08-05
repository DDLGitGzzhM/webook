package repository

import (
	"context"
	"time"

	"webook/webook/account/domain"
	"webook/webook/account/repository/cache"
	"webook/webook/account/repository/dao"
)

type accountRepository struct {
	dao   dao.AccountDAO
	cache cache.Cache
}

// NewAccountRepository 创建账户仓储
func NewAccountRepository(dao dao.AccountDAO, cache cache.Cache) AccountRepository {
	return &accountRepository{dao: dao, cache: cache}
}

// CheckUnique 如果返回了 error 就说明重复记账了
func (a *accountRepository) CheckUnique(ctx context.Context, c domain.Credit) error {
	return a.cache.GetUnique(ctx, c)
}

func (a *accountRepository) SetUnique(ctx context.Context, c domain.Credit) error {
	return a.cache.SetUnique(ctx, c)
}

func (a *accountRepository) AddCredit(ctx context.Context, c domain.Credit) error {
	activities := make([]dao.AccountActivity, 0, len(c.Items))
	now := time.Now().UnixMilli()
	for _, itm := range c.Items {
		activities = append(activities, dao.AccountActivity{
			Uid:         itm.Uid,
			Biz:         c.Biz,
			BizId:       c.BizId,
			Account:     itm.Account,
			AccountType: itm.AccountType.AsUint8(),
			Amount:      itm.Amt,
			Currency:    itm.Currency,
			Ctime:       now,
			Utime:       now,
		})
	}
	// 把它改成了记录账号变动活动，同时会去更新余额
	return a.dao.AddActivities(ctx, activities...)
}
