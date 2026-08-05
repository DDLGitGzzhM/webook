//go:build wireinject

package startup

import (
	"github.com/google/wire"

	"webook/webook/follow/grpc"
	"webook/webook/follow/repository"
	"webook/webook/follow/repository/cache"
	"webook/webook/follow/repository/dao"
	"webook/webook/follow/service"
	"webook/webook/internal/pkg/logger"
)

func InitServer() *grpc.FollowServiceServer {
	wire.Build(
		InitRedis,
		logger.NewNoOpLogger,
		InitTestDB,
		dao.NewGORMFollowRelationDAO,
		cache.NewRedisFollowCache,
		repository.NewFollowRelationRepository,
		service.NewFollowRelationService,
		grpc.NewFollowRelationServiceServer,
	)
	return new(grpc.FollowServiceServer)
}
