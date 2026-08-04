//go:build wireinject

package main

import (
	"github.com/google/wire"

	"webook/webook/account/grpc"
	"webook/webook/account/ioc"
	"webook/webook/account/repository"
	"webook/webook/account/repository/dao"
	"webook/webook/account/service"
)

func Init() *App {
	wire.Build(
		ioc.InitDB,
		ioc.InitLogger,
		ioc.InitGRPCxServer,
		dao.NewCreditGORMDAO,
		repository.NewAccountRepository,
		service.NewAccountService,
		grpc.NewAccountServiceServer,
		wire.Struct(new(App), "GRPCServer"))
	return new(App)
}
