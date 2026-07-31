package ioc

import (
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	grpc2 "webook/webook/interactive/grpc"
	"webook/webook/pkg/grpcx"
)

func InitGRPCxServer(intrServer *grpc2.InteractiveServiceServer) *grpcx.Server {
	type Config struct {
		Addr string `yaml:"addr"`
	}

	var cfg Config
	err := viper.UnmarshalKey("grpc.server", &cfg)
	if err != nil {
		panic(err)
	}

	server := grpc.NewServer()
	intrServer.Register(server)

	return &grpcx.Server{
		Server: server,
		Addr:   cfg.Addr,
	}
}
