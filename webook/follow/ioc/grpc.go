package ioc

import (
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	grpc2 "webook/webook/follow/grpc"
	"webook/webook/internal/pkg/logger"
	"webook/webook/pkg/grpcx"
)

// InitGRPCxServer 初始化 gRPC 服务并注册到 etcd
func InitGRPCxServer(
	followRelation *grpc2.FollowServiceServer,
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
	followRelation.Register(server)
	return &grpcx.Server{
		Server:    server,
		Port:      cfg.Port,
		Name:      "follow",
		L:         l,
		EtcdAddrs: cfg.EtcdAddrs,
	}
}
