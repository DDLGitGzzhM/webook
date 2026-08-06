//go:build wireinject

package main

import (
	"github.com/google/wire"

	"webook/webook/search/events"
	"webook/webook/search/grpc"
	"webook/webook/search/ioc"
	"webook/webook/search/repository"
	"webook/webook/search/repository/dao"
	"webook/webook/search/service"
)

var serviceProviderSet = wire.NewSet(
	dao.NewUserElasticDAO,
	dao.NewArticleElasticDAO,
	dao.NewAnyESDAO,
	dao.NewTagESDAO,
	repository.NewUserRepository,
	repository.NewArticleRepository,
	repository.NewAnyRepository,
	service.NewSyncService,
	service.NewSearchService,
)

var thirdProvider = wire.NewSet(
	ioc.InitESClient,
	ioc.InitLogger,
	ioc.InitKafka,
)

func Init() *App {
	wire.Build(
		thirdProvider,
		serviceProviderSet,
		grpc.NewSyncServiceServer,
		grpc.NewSearchService,
		events.NewUserConsumer,
		events.NewArticleConsumer,
		events.NewSyncDataEventConsumer,
		ioc.InitGRPCxServer,
		ioc.NewConsumers,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
