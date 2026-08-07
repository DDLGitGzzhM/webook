package ioc

import (
	followv1 "webook/webook/api/proto/gen/follow/v1"
	"webook/webook/feed/repository"
	"webook/webook/feed/service"
)

func RegisterHandler(repo repository.FeedEventRepo, followClient followv1.FollowServiceClient) map[string]service.Handler {
	articleHandler := service.NewArticleEventHandler(repo, followClient)
	followHanlder := service.NewFollowEventHandler(repo)
	likeHandler := service.NewLikeEventHandler(repo)
	return map[string]service.Handler{
		service.ArticleEventName: articleHandler,
		service.FollowEventName:  followHanlder,
		service.LikeEventName:    likeHandler,
	}
}
