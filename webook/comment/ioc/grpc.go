package ioc

import (
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	grpc2 "webook/webook/comment/grpc"
	"webook/webook/internal/pkg/logger"
	"webook/webook/pkg/grpcx"
)

func InitGRPCxServer(
	comment *grpc2.CommentServiceServer,
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
	comment.Register(server)
	return &grpcx.Server{
		Server:    server,
		Port:      cfg.Port,
		Name:      "comment",
		L:         l,
		EtcdAddrs: cfg.EtcdAddrs,
	}
}
