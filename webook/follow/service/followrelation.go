package service

import (
	"context"

	"webook/webook/follow/domain"
	"webook/webook/follow/repository"
)

// FollowRelationService 关注关系服务
type FollowRelationService interface {
	GetFollowee(ctx context.Context, follower, offset, limit int64) ([]domain.FollowRelation, error)
	GetFollower(ctx context.Context, followee, offset, limit int64) ([]domain.FollowRelation, error)
	FollowInfo(ctx context.Context,
		follower, followee int64) (domain.FollowRelation, error)
	Follow(ctx context.Context, follower, followee int64) error
	CancelFollow(ctx context.Context, follower, followee int64) error
	GetFollowStatic(ctx context.Context, uid int64) (domain.FollowStatics, error)
}

type followRelationService struct {
	repo repository.FollowRepository
}

func (f *followRelationService) CancelFollow(ctx context.Context, follower, followee int64) error {
	return f.repo.InactiveFollowRelation(ctx, follower, followee)
}

// NewFollowRelationService 创建关注关系服务
func NewFollowRelationService(repo repository.FollowRepository) FollowRelationService {
	return &followRelationService{
		repo: repo,
	}
}

func (f *followRelationService) GetFollowee(ctx context.Context,
	follower, offset, limit int64) ([]domain.FollowRelation, error) {
	return f.repo.GetFollowee(ctx, follower, offset, limit)
}

func (f *followRelationService) GetFollower(ctx context.Context,
	followee, offset, limit int64) ([]domain.FollowRelation, error) {
	return f.repo.GetFollower(ctx, followee, offset, limit)
}

func (f *followRelationService) GetFollowStatic(ctx context.Context, uid int64) (domain.FollowStatics, error) {
	return f.repo.GetFollowStatics(ctx, uid)
}

func (f *followRelationService) FollowInfo(ctx context.Context, follower, followee int64) (domain.FollowRelation, error) {
	val, err := f.repo.FollowInfo(ctx, follower, followee)
	return val, err
}

func (f *followRelationService) Follow(ctx context.Context, follower, followee int64) error {
	return f.repo.AddFollowRelation(ctx, domain.FollowRelation{
		Followee: followee,
		Follower: follower,
	})
}
