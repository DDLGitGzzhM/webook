package repository

import (
	"context"
	"encoding/json"
	"time"

	"webook/webook/feed/domain"
	"webook/webook/feed/repository/cache"
	"webook/webook/feed/repository/dao"
)

var FolloweesNotFound = cache.FolloweesNotFound

type FeedEventRepo interface {
	// CreatePushEvents 批量推事件
	CreatePushEvents(ctx context.Context, events []domain.FeedEvent) error
	// CreatePullEvent 创建拉事件
	CreatePullEvent(ctx context.Context, event domain.FeedEvent) error
	// FindPullEvents 获取拉事件，也就是关注的人发件箱里面的事件
	FindPullEvents(ctx context.Context, uids []int64, timestamp, limit int64) ([]domain.FeedEvent, error)
	// FindPushEvents 获取推事件，也就是自己收件箱里面的事件
	FindPushEvents(ctx context.Context, uid, timestamp, limit int64) ([]domain.FeedEvent, error)
}

type feedEventRepo struct {
	pullDao dao.FeedPullEventDAO
	pushDao dao.FeedPushEventDAO
}

func NewFeedEventRepo(pullDao dao.FeedPullEventDAO, pushDao dao.FeedPushEventDAO, feedCache cache.FeedEventCache) FeedEventRepo {
	return &feedEventRepo{
		pullDao: pullDao,
		pushDao: pushDao,
	}
}

func (f *feedEventRepo) FindPullEvents(ctx context.Context, uids []int64, timestamp, limit int64) ([]domain.FeedEvent, error) {
	panic("implement me")
}

func (f *feedEventRepo) FindPushEvents(ctx context.Context, uid, timestamp, limit int64) ([]domain.FeedEvent, error) {
	panic("implement me")
}

func (f *feedEventRepo) CreatePushEvents(ctx context.Context, events []domain.FeedEvent) error {
	pushEvents := make([]dao.FeedPushEvent, 0, len(events))
	for _, e := range events {
		pushEvents = append(pushEvents, convertToPushEventDao(e))
	}
	return f.pushDao.CreatePushEvents(ctx, pushEvents)
}

func (f *feedEventRepo) CreatePullEvent(ctx context.Context, event domain.FeedEvent) error {
	return f.pullDao.CreatePullEvent(ctx, convertToPullEventDao(event))
}

func convertToPushEventDao(event domain.FeedEvent) dao.FeedPushEvent {
	val, _ := json.Marshal(event.Ext)
	return dao.FeedPushEvent{
		Id:      event.ID,
		UID:     event.Uid,
		Type:    event.Type,
		Content: string(val),
		Ctime:   event.Ctime.Unix(),
	}
}

func convertToPullEventDao(event domain.FeedEvent) dao.FeedPullEvent {
	val, _ := json.Marshal(event.Ext)
	return dao.FeedPullEvent{
		Id:      event.ID,
		UID:     event.Uid,
		Type:    event.Type,
		Content: string(val),
		Ctime:   event.Ctime.Unix(),
	}

}

func convertToPushEventDomain(event dao.FeedPushEvent) domain.FeedEvent {
	var ext map[string]string
	_ = json.Unmarshal([]byte(event.Content), &ext)
	return domain.FeedEvent{
		ID:    event.Id,
		Uid:   event.UID,
		Type:  event.Type,
		Ctime: time.Unix(event.Ctime, 0),
		Ext:   ext,
	}
}

func convertToPullEventDomain(event dao.FeedPullEvent) domain.FeedEvent {
	var ext map[string]string
	_ = json.Unmarshal([]byte(event.Content), &ext)
	return domain.FeedEvent{
		ID:    event.Id,
		Uid:   event.UID,
		Type:  event.Type,
		Ctime: time.Unix(event.Ctime, 0),
		Ext:   ext,
	}
}
