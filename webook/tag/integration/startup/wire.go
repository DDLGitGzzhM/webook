//go:build wireinject

package startup

import (
	"github.com/google/wire"

	"webook/webook/tag/events"
	"webook/webook/tag/grpc"
	"webook/webook/tag/repository/cache"
	"webook/webook/tag/repository/dao"
	"webook/webook/tag/service"
)

// InitGRPCService 初始化集成测试用的 Tag gRPC 服务。
func InitGRPCService(p events.Producer) *grpc.TagServiceServer {
	wire.Build(
		InitTestDB,
		InitRedis,
		InitLog,
		dao.NewGORMTagDAO,
		InitRepository,
		cache.NewRedisTagCache,
		service.NewTagService,
		grpc.NewTagServiceServer,
	)
	return new(grpc.TagServiceServer)
}
