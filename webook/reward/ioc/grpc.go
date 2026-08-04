package ioc

import (
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	"webook/webook/internal/pkg/logger"
	"webook/webook/pkg/grpcx"
	grpc2 "webook/webook/reward/grpc"
)

func InitGRPCxServer(
	reward *grpc2.RewardServiceServer,
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
	reward.Register(server)
	return &grpcx.Server{
		Server:    server,
		Port:      cfg.Port,
		Name:      "reward",
		L:         l,
		EtcdAddrs: cfg.EtcdAddrs,
	}
}
