package repository

import (
	"context"

	"webook/webook/internal/domain"
	"webook/webook/internal/repository/cache"
	"webook/webook/internal/repository/dao"
)

var ErrUserDuplicateEmail = dao.ErrUserDuplicateEmail
var ErrUserNotFound = dao.ErrUserNotFound

type UserRepository struct {
	dao   *dao.UserDAO
	cache *cache.UserCache
}

func NewUserRepository(dao *dao.UserDAO, cache *cache.UserCache) *UserRepository {
	return &UserRepository{
		dao:   dao,
		cache: cache,
	}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	return r.dao.Insert(ctx, &dao.User{
		Email:    user.Email,
		Password: user.PassWord,
	})
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := r.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return domain.User{
		Id:       u.Id,
		Email:    u.Email,
		PassWord: u.Password,
	}, nil
}

func (r *UserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	u, err := r.cache.GetUser(ctx, id)
	if err == nil {
		return u, nil
	}
	ue, err := r.dao.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}
	u = domain.User{
		Id:       ue.Id,
		Email:    ue.Email,
		PassWord: ue.Password,
	}
	err = r.cache.SetUser(ctx, u)
	if err != nil {
		//打日志做监控
		//return domain.User{}, err
	}
	return u, nil
	/*
			如果 redis 整个都崩了，数据库会接受大量流量 数据库 也会崩
		    这里有两个决策
			1. 选加载数据库 --- 做好兜底，万一 Redis 真的崩了，你要保护住你的数据库
				- 数据库限流
			2. 选不加载 --- 用户体验差一点
	*/
}
