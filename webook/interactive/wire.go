//go:build wireinject

package main

import (
	"github.com/google/wire"

	"webook/webook/interactive/grpc"
	"webook/webook/interactive/ioc"
	"webook/webook/internal/events/article"
	"webook/webook/internal/repository"
	"webook/webook/internal/repository/cache"
	"webook/webook/internal/repository/dao"
	"webook/webook/internal/service"
)

var thirdPartySet = wire.NewSet(
	ioc.InitDST,
	ioc.InitSRC,
	ioc.InitBizDB,
	ioc.InitDoubleWritePool,
	ioc.InitLogger,
	ioc.InitKafka,
	ioc.InitSyncProducer,
	ioc.InitRedis,
)

var interactiveSvcProvider = wire.NewSet(
	service.NewInteractiveService,
	repository.NewCachedInteractiveRepository,
	dao.NewGORMInteractiveDAO,
	cache.NewRedisInteractiveCache,
)

var migratorProvider = wire.NewSet(
	ioc.InitMigratorWeb,
	ioc.InitFixDataConsumer,
	ioc.InitMigradatorProducer,
	ioc.InitMySQLBinlogConsumer,
)

func InitAPP() *App {
	wire.Build(
		interactiveSvcProvider,
		thirdPartySet,
		migratorProvider,
		article.NewInteractiveReadEventBatchConsumer,
		grpc.NewInteractiveServiceServer,
		ioc.NewConsumers,
		ioc.InitGRPCxServer,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
