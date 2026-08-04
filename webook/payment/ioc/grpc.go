package ioc

import (
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	"webook/webook/internal/pkg/logger"
	grpc2 "webook/webook/payment/grpc"
	"webook/webook/pkg/grpcx"
)

func InitGRPCServer(
	wesvc *grpc2.WechatServiceServer,
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
	wesvc.Register(server)
	return &grpcx.Server{
		Server:    server,
		Port:      cfg.Port,
		Name:      "payment",
		L:         l,
		EtcdAddrs: cfg.EtcdAddrs,
	}
}
