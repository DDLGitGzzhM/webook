package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"webook/webook/follow/domain"
)

type fakeFollowRepo struct {
	followees []domain.FollowRelation
	info      domain.FollowRelation
	infoErr   error
	addErr    error
	inactive  error
	added     []domain.FollowRelation
	canceled  [][2]int64
}

func (f *fakeFollowRepo) GetFollowee(ctx context.Context, follower, offset, limit int64) ([]domain.FollowRelation, error) {
	return f.followees, nil
}

func (f *fakeFollowRepo) FollowInfo(ctx context.Context, follower int64, followee int64) (domain.FollowRelation, error) {
	return f.info, f.infoErr
}

func (f *fakeFollowRepo) AddFollowRelation(ctx context.Context, fr domain.FollowRelation) error {
	f.added = append(f.added, fr)
	return f.addErr
}

func (f *fakeFollowRepo) InactiveFollowRelation(ctx context.Context, follower int64, followee int64) error {
	f.canceled = append(f.canceled, [2]int64{follower, followee})
	return f.inactive
}

func (f *fakeFollowRepo) GetFollowStatics(ctx context.Context, uid int64) (domain.FollowStatics, error) {
	return domain.FollowStatics{}, nil
}

func TestFollowRelationService_Follow(t *testing.T) {
	repo := &fakeFollowRepo{}
	svc := NewFollowRelationService(repo)
	err := svc.Follow(context.Background(), 1, 2)
	require.NoError(t, err)
	require.Len(t, repo.added, 1)
	assert.Equal(t, domain.FollowRelation{Follower: 1, Followee: 2}, repo.added[0])
}

func TestFollowRelationService_CancelFollow(t *testing.T) {
	repo := &fakeFollowRepo{}
	svc := NewFollowRelationService(repo)
	err := svc.CancelFollow(context.Background(), 1, 2)
	require.NoError(t, err)
	require.Len(t, repo.canceled, 1)
	assert.Equal(t, [2]int64{1, 2}, repo.canceled[0])
}

func TestFollowRelationService_FollowInfo(t *testing.T) {
	want := domain.FollowRelation{Follower: 1, Followee: 2}
	repo := &fakeFollowRepo{info: want}
	svc := NewFollowRelationService(repo)
	got, err := svc.FollowInfo(context.Background(), 1, 2)
	require.NoError(t, err)
	assert.Equal(t, want, got)

	repo.infoErr = errors.New("not found")
	_, err = svc.FollowInfo(context.Background(), 1, 2)
	assert.Error(t, err)
}

func TestFollowRelationService_GetFollowee(t *testing.T) {
	want := []domain.FollowRelation{{Follower: 1, Followee: 2}}
	repo := &fakeFollowRepo{followees: want}
	svc := NewFollowRelationService(repo)
	got, err := svc.GetFollowee(context.Background(), 1, 0, 10)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}
