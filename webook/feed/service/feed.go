package service

import (
	"context"
	"fmt"

	followv1 "webook/webook/api/proto/gen/follow/v1"
	"webook/webook/feed/domain"
	"webook/webook/feed/repository"
)

type feedService struct {

	// key 就是 type，value 具体的业务处理逻辑
	handlerMap map[string]Handler

	defaultHdl Handler

	repo         repository.FeedEventRepo
	followClient followv1.FollowServiceClient
}

// NewFeedService 在 IOC 完成组装
func NewFeedService(repo repository.FeedEventRepo, handlerMap map[string]Handler) FeedService {
	return &feedService{
		repo:       repo,
		handlerMap: handlerMap,
	}
}

func (f *feedService) Register(typ string, hdl Handler) {
	f.handlerMap[typ] = hdl
}

func (f *feedService) CreateFeedEvent(ctx context.Context, feed domain.FeedEvent) error {
	handler, ok := f.handlerMap[feed.Type]
	if !ok {
		// 这里，基本上就是代码错误，或者业务方传递过来的参数错误
		// 还有另外一种做法，就是走兜底路径
		//return f.defaultHdl.CreateFeedEvent(ctx, feed.Ext)
		return fmt.Errorf("未找到正确的业务 handler %s", feed.Type)
	}
	return handler.CreateFeedEvent(ctx, feed.Ext)
}

// GetFeedEventListV1 不依赖于 Handler 的直接查询
func (f *feedService) GetFeedEventListV1(ctx context.Context, uid int64, timestamp, limit int64) ([]domain.FeedEvent, error) {
	panic("implement me")
}

func (f *feedService) GetFeedEventList(ctx context.Context, uid int64, timestamp, limit int64) ([]domain.FeedEvent, error) {
	panic("implement me")
}
