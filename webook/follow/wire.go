//go:build wireinject

package main

import (
	"github.com/google/wire"

	"webook/webook/follow/grpc"
	"webook/webook/follow/ioc"
	"webook/webook/follow/repository"
	"webook/webook/follow/repository/cache"
	"webook/webook/follow/repository/dao"
	"webook/webook/follow/service"
)

var serviceProviderSet = wire.NewSet(
	dao.NewGORMFollowRelationDAO,
	cache.NewRedisFollowCache,
	repository.NewFollowRelationRepository,
	service.NewFollowRelationService,
	grpc.NewFollowRelationServiceServer,
)

var thirdProvider = wire.NewSet(
	ioc.InitDB,
	ioc.InitRedis,
	ioc.InitLogger,
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
