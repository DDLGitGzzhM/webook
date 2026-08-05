//go:build wireinject

package startup

import (
	"github.com/google/wire"

	grpc2 "webook/webook/comment/grpc"
	"webook/webook/comment/repository"
	"webook/webook/comment/repository/dao"
	"webook/webook/comment/service"
	"webook/webook/internal/pkg/logger"
)

var serviceProviderSet = wire.NewSet(
	dao.NewCommentDAO,
	repository.NewCommentRepo,
	service.NewCommentSvc,
	grpc2.NewGrpcServer,
)

var thirdProvider = wire.NewSet(
	logger.NewNoOpLogger,
	InitTestDB,
)

func InitGRPCServer() *grpc2.CommentServiceServer {
	wire.Build(thirdProvider, serviceProviderSet)
	return new(grpc2.CommentServiceServer)
}
