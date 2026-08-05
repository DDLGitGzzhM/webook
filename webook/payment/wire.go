//go:build wireinject

package main

import (
	"github.com/google/wire"

	"webook/webook/payment/grpc"
	"webook/webook/payment/ioc"
	"webook/webook/payment/repository"
	"webook/webook/payment/repository/dao"
	"webook/webook/payment/web"
)

func InitApp() *App {
	wire.Build(
		ioc.InitKafka,
		ioc.InitProducer,
		ioc.InitWechatClient,
		dao.NewPaymentGORMDAO,
		ioc.InitDB,
		repository.NewPaymentRepository,
		repository.NewPaymentGORMRepository,
		repository.NewLocalMsgGORMRepository,
		grpc.NewWechatServiceServer,
		ioc.InitWechatNativeService,
		ioc.InitWechatConfig,
		ioc.InitWechatNotifyHandler,
		ioc.InitGRPCServer,
		web.NewWechatHandler,
		ioc.InitGinServer,
		ioc.InitLogger,
		wire.Struct(new(App), "WebServer", "GRPCServer"))
	return new(App)
}
