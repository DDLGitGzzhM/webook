//go:build wireinject

package main

import (
	"github.com/google/wire"

	"webook/webook/feed/events"
	"webook/webook/feed/grpc"
	"webook/webook/feed/ioc"
	"webook/webook/feed/repository"
	"webook/webook/feed/repository/cache"
	"webook/webook/feed/repository/dao"
	"webook/webook/feed/service"
)

var serviceProviderSet = wire.NewSet(
	dao.NewFeedPushEventDAO,
	dao.NewFeedPullEventDAO,
	cache.NewFeedEventCache,
	repository.NewFeedEventRepo,
)

var thirdProvider = wire.NewSet(
	ioc.InitEtcdClient,
	ioc.InitLogger,
	ioc.InitRedis,
	ioc.InitKafka,
	ioc.InitDB,
	ioc.InitFollowClient,
)

func Init() *App {
	wire.Build(
		thirdProvider,
		serviceProviderSet,
		ioc.RegisterHandler,
		service.NewFeedService,
		grpc.NewFeedEventGrpcSvc,
		events.NewArticleEventConsumer,
		events.NewFeedEventConsumer,
		ioc.InitGRPCxServer,
		ioc.NewConsumers,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
