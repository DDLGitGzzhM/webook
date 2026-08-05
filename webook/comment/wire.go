//go:build wireinject

package main

import (
	"github.com/google/wire"

	"webook/webook/comment/grpc"
	"webook/webook/comment/ioc"
	"webook/webook/comment/repository"
	"webook/webook/comment/repository/dao"
	"webook/webook/comment/service"
)

var serviceProviderSet = wire.NewSet(
	dao.NewCommentDAO,
	repository.NewCommentRepo,
	service.NewCommentSvc,
	grpc.NewGrpcServer,
)

var thirdProvider = wire.NewSet(
	ioc.InitLogger,
	ioc.InitDB,
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
