package ioc

import (
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	"webook/webook/internal/pkg/logger"
	"webook/webook/pkg/grpcx"
	grpc2 "webook/webook/tag/grpc"
)

// InitGRPCxServer 初始化 gRPC 服务并注册到 etcd。
func InitGRPCxServer(
	tagSvc *grpc2.TagServiceServer,
	l logger.Logger,
) *grpcx.Server {
	type Config struct {
		Port      int      `yaml:"port"`
		EtcdAddrs []string `yaml:"etcdAddrs"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.server", &cfg)
	if err != nil {
		panic(err)
	}
	server := grpc.NewServer()
	tagSvc.Register(server)
	return &grpcx.Server{
		Server:    server,
		Port:      cfg.Port,
		Name:      "tag",
		L:         l,
		EtcdAddrs: cfg.EtcdAddrs,
	}
}
