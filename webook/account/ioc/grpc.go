package ioc

import (
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	grpc3 "webook/webook/account/grpc"
	"webook/webook/internal/pkg/logger"
	"webook/webook/pkg/grpcx"
)

func InitGRPCxServer(
	asc *grpc3.AccountServiceServer,
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
	asc.Register(server)
	return &grpcx.Server{
		Server:    server,
		Port:      cfg.Port,
		Name:      "account",
		L:         l,
		EtcdAddrs: cfg.EtcdAddrs,
	}
}
