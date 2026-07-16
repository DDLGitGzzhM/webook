package repository

import (
	"context"
	"database/sql"
	"time"

	"webook/webook/internal/domain"
	"webook/webook/internal/repository/cache"
	"webook/webook/internal/repository/dao"
)

var ErrUserDuplicateEmail = dao.ErrUserDuplicateEmail
var ErrUserNotFound = dao.ErrUserNotFound

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
	FindById(ctx context.Context, id int64) (domain.User, error)
}
type CacheUserRepository struct {
	dao   dao.UserDao
	cache cache.UserCache
}

func NewUserRepository(dao dao.UserDao, cache cache.UserCache) *CacheUserRepository {
	return &CacheUserRepository{
		dao:   dao,
		cache: cache,
	}
}

func (r *CacheUserRepository) Create(ctx context.Context, user *domain.User) error {
	return r.dao.Insert(ctx, r.domainToEntity(*user))
}

func (r *CacheUserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := r.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(u), nil
}

func (r *CacheUserRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	u, err := r.dao.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(u), nil
}

func (r *CacheUserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	u, err := r.cache.GetUser(ctx, id)
	if err == nil {
		return u, nil
	}
	ue, err := r.dao.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}
	u = r.entityToDomain(ue)
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

func (r *CacheUserRepository) domainToEntity(u domain.User) *dao.User {
	return &dao.User{
		Id:       u.Id,
		Email:    sql.NullString{String: u.Email, Valid: true},
		Phone:    sql.NullString{String: u.Phone, Valid: true},
		Password: u.PassWord,
		Ctime:    u.Ctime.UnixMilli(),
	}
}

func (r *CacheUserRepository) entityToDomain(u *dao.User) domain.User {
	return domain.User{
		Id:       u.Id,
		Email:    u.Email.String,
		Phone:    u.Phone.String,
		PassWord: u.Password,
		Ctime:    time.UnixMilli(u.Ctime),
	}
}
