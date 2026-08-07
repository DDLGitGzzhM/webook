package service

import (
	"context"

	"github.com/ecodeclub/ekit/slice"

	followv1 "webook/webook/api/proto/gen/follow/v1"
	"webook/webook/feed/domain"
	"webook/webook/feed/repository"
)

type ArticleEventHandler struct {
	repo         repository.FeedEventRepo
	followClient followv1.FollowServiceClient
}

const (
	ArticleEventName = "article_event"
	// 你可以调大或者调小
	// 调大，数据量大，但是用户体验好
	// 调小，数据量小，但是用户体验差
	threshold = 32
)

func NewArticleEventHandler(repo repository.FeedEventRepo, client followv1.FollowServiceClient) Handler {
	return &ArticleEventHandler{
		repo:         repo,
		followClient: client,
	}
}

func (a *ArticleEventHandler) FindFeedEvents(ctx context.Context, uid, timestamp, limit int64) ([]domain.FeedEvent, error) {
	panic("implement me")
}

func (a *ArticleEventHandler) CreateFeedEvent(ctx context.Context, ext domain.ExtendFields) error {
	followee, err := ext.Get("uid").AsInt64()
	if err != nil {
		return err
	}
	// 要灵活判定是拉模型（读扩散）还是推模型（写扩散）
	static, err := a.followClient.GetFollowStatic(ctx, &followv1.GetFollowStaticRequest{
		Followee: followee,
	})
	if err != nil {
		return err
	}
	// 粉丝数超过阈值了，然后读扩散，不然写扩散
	if static.FollowStatic.Followers > threshold {
		return a.repo.CreatePullEvent(ctx, domain.FeedEvent{
			Type: ArticleEventName,
			Uid:  followee,
			Ext:  ext,
		})
	} else {
		// 写扩散
		followers, err := a.followClient.GetFollower(ctx, &followv1.GetFollowerRequest{Followee: followee})
		if err != nil {
			return err
		}
		// 在这里，判定写扩散还是读扩散
		// 要综合考虑什么活跃用户，是不是铁粉，
		// 在这里判定
		events := slice.Map(followers.FollowRelations,
			func(idx int, src *followv1.FollowRelation) domain.FeedEvent {
				return domain.FeedEvent{
					Uid:  src.Follower,
					Type: ArticleEventName,
					Ext:  ext,
				}
			})
		return a.repo.CreatePushEvents(ctx, events)
	}
}
