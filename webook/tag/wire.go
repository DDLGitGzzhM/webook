//go:build wireinject

package main

import (
	"github.com/google/wire"

	"webook/webook/tag/events"
	"webook/webook/tag/grpc"
	"webook/webook/tag/ioc"
	"webook/webook/tag/repository/cache"
	"webook/webook/tag/repository/dao"
	"webook/webook/tag/service"
)

var serviceProviderSet = wire.NewSet(
	dao.NewGORMTagDAO,
	cache.NewRedisTagCache,
	ioc.InitRepository,
	events.NewSaramaSyncProducer,
	service.NewTagService,
	grpc.NewTagServiceServer,
)

var thirdProvider = wire.NewSet(
	ioc.InitDB,
	ioc.InitRedis,
	ioc.InitLogger,
	ioc.InitKafka,
	ioc.InitSyncProducer,
)

func Init() *App {
	wire.Build(
		thirdProvider,
		serviceProviderSet,
		ioc.InitGRPCxServer,
		wire.Struct(new(App), "GRPCServer"),
	)
	return new(App)
}
