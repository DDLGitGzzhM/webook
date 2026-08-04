//go:build wireinject

package main

import (
	"github.com/google/wire"

	"webook/webook/reward/grpc"
	"webook/webook/reward/ioc"
	"webook/webook/reward/repository"
	"webook/webook/reward/repository/cache"
	"webook/webook/reward/repository/dao"
	"webook/webook/reward/service"
)

var thirdPartySet = wire.NewSet(
	ioc.InitDB,
	ioc.InitLogger,
	ioc.InitEtcdClient,
	ioc.InitRedis,
	ioc.InitKafka,
)

func Init() *App {
	wire.Build(thirdPartySet,
		service.NewWechatNativeRewardService,
		wire.Bind(new(service.RewardService), new(*service.WechatNativeRewardService)),
		ioc.InitAccountClient,
		ioc.InitGRPCxServer,
		ioc.InitPaymentClient,
		ioc.InitPaymentEventConsumer,
		repository.NewRewardRepository,
		cache.NewRewardRedisCache,
		dao.NewRewardGORMDAO,
		grpc.NewRewardServiceServer,
		wire.Struct(new(App), "GRPCServer", "Consumer"),
	)
	return new(App)
}
