//go:build wireinject

package startup

import (
	"github.com/google/wire"

	"webook/webook/account/grpc"
	"webook/webook/account/repository"
	"webook/webook/account/repository/cache"
	"webook/webook/account/repository/dao"
	"webook/webook/account/service"
)

func InitAccountService() *grpc.AccountServiceServer {
	wire.Build(
		InitTestDB,
		InitRedis,
		dao.NewCreditGORMDAO,
		cache.NewRedisCache,
		repository.NewAccountRepository,
		service.NewAccountService,
		grpc.NewAccountServiceServer,
	)
	return new(grpc.AccountServiceServer)
}
