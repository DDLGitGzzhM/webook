package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"webook/webook/feed/domain"
)

type fakeFeedRepo struct {
	pushEvents []domain.FeedEvent
	pullEvent  domain.FeedEvent
}

func (f *fakeFeedRepo) CreatePushEvents(ctx context.Context, events []domain.FeedEvent) error {
	f.pushEvents = append(f.pushEvents, events...)
	return nil
}

func (f *fakeFeedRepo) CreatePullEvent(ctx context.Context, event domain.FeedEvent) error {
	f.pullEvent = event
	return nil
}

func (f *fakeFeedRepo) FindPullEvents(ctx context.Context, uids []int64, timestamp, limit int64) ([]domain.FeedEvent, error) {
	return nil, nil
}

func (f *fakeFeedRepo) FindPushEvents(ctx context.Context, uid, timestamp, limit int64) ([]domain.FeedEvent, error) {
	return nil, nil
}

func (f *fakeFeedRepo) FindPullEventsWithTyp(ctx context.Context, typ string, uids []int64, timestamp, limit int64) ([]domain.FeedEvent, error) {
	return nil, nil
}

func (f *fakeFeedRepo) FindPushEventsWithTyp(ctx context.Context, typ string, uid, timestamp, limit int64) ([]domain.FeedEvent, error) {
	return nil, nil
}

func TestLikeEventHandler_CreateFeedEvent(t *testing.T) {
	repo := &fakeFeedRepo{}
	hdl := NewLikeEventHandler(repo)
	err := hdl.CreateFeedEvent(context.Background(), domain.ExtendFields{
		"liked": "11",
		"liker": "22",
	})
	require.NoError(t, err)
	require.Len(t, repo.pushEvents, 1)
	assert.Equal(t, int64(11), repo.pushEvents[0].Uid)
	assert.Equal(t, LikeEventName, repo.pushEvents[0].Type)
}

func TestFollowEventHandler_CreateFeedEvent(t *testing.T) {
	repo := &fakeFeedRepo{}
	hdl := NewFollowEventHandler(repo)
	err := hdl.CreateFeedEvent(context.Background(), domain.ExtendFields{
		"followee": "33",
		"follower": "44",
	})
	require.NoError(t, err)
	require.Len(t, repo.pushEvents, 1)
	assert.Equal(t, int64(33), repo.pushEvents[0].Uid)
	assert.Equal(t, FollowEventName, repo.pushEvents[0].Type)
}

func TestFeedService_CreateFeedEvent_UnknownType(t *testing.T) {
	svc := NewFeedService(&fakeFeedRepo{}, map[string]Handler{})
	err := svc.CreateFeedEvent(context.Background(), domain.FeedEvent{Type: "unknown"})
	assert.Error(t, err)
}
