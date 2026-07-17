package service

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"webook/webook/internal/domain"
	"webook/webook/internal/repository"
)

var ErrUserDuplicateEmail = repository.ErrUserDuplicateEmail
var ErrInvalidPassword = errors.New("邮箱/密码不对")

type IUserService interface {
	SignUp(ctx context.Context, u domain.User) error
	Login(ctx context.Context, email string, password string) (domain.User, error)
	FindOrCreate(ctx context.Context, phone string) (domain.User, error)
	FindOrCreateByWechat(ctx context.Context, weChatInfo domain.WeChatInfo) (domain.User, error)
	Profile(ctx context.Context, id int64) (domain.User, error)
}
type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (svc *UserService) SignUp(ctx context.Context, u domain.User) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.PassWord), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PassWord = string(hash)
	return svc.repo.Create(ctx, &u)
}

func (svc *UserService) Login(ctx context.Context, email string, password string) (domain.User, error) {
	u, err := svc.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return u, ErrInvalidPassword
		}
		return u, err
	}
	if err = bcrypt.CompareHashAndPassword([]byte(u.PassWord), []byte(password)); err != nil {
		return u, ErrInvalidPassword
	}
	return u, nil
}

func (svc *UserService) FindOrCreateByWechat(ctx context.Context, weChatInfo domain.WeChatInfo) (domain.User, error) {
	u, err := svc.repo.FindByWechat(ctx, weChatInfo.OpenId)
	if !errors.Is(err, repository.ErrUserNotFound) {
		return u, err
	}
	u = domain.User{
		WeChatInfo: weChatInfo,
	}
	err = svc.repo.Create(ctx, &u)
	if err != nil {
		return domain.User{}, err
	}
	u, err = svc.repo.FindByWechat(ctx, weChatInfo.OpenId)
	if err != nil {
		return u, err
	} // 这里会遇到主从延迟的问题
	return u, nil
}

func (svc *UserService) FindOrCreate(ctx context.Context, phone string) (domain.User, error) {
	u, err := svc.repo.FindByPhone(ctx, phone)
	if !errors.Is(err, repository.ErrUserNotFound) {
		return u, err
	}
	err = svc.repo.Create(ctx, &domain.User{
		Phone: phone,
	})
	if err != nil {
		return domain.User{}, err
	}
	u, err = svc.repo.FindByPhone(ctx, phone)
	if err != nil {
		return u, err
	} // 这里会遇到主从延迟的问题
	return u, nil
}

func (svc *UserService) Profile(ctx context.Context, id int64) (domain.User, error) {
	u, err := svc.repo.FindById(ctx, id)
	if err != nil {
		return u, err
	}
	return u, nil
}
