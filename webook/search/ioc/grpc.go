package ioc

import (
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	"webook/webook/internal/pkg/logger"
	"webook/webook/pkg/grpcx"
	grpc2 "webook/webook/search/grpc"
)

// InitGRPCxServer 初始化 gRPC 服务并注册到 etcd。
func InitGRPCxServer(
	syncRpc *grpc2.SyncServiceServer,
	searchRpc *grpc2.SearchServiceServer,
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
	syncRpc.Register(server)
	searchRpc.Register(server)
	return &grpcx.Server{
		Server:    server,
		Port:      cfg.Port,
		Name:      "search",
		L:         l,
		EtcdAddrs: cfg.EtcdAddrs,
	}
}
