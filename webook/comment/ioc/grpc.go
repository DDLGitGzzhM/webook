package ioc

import (
	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"

	grpc2 "webook/webook/comment/grpc"
	"webook/webook/internal/pkg/logger"
	"webook/webook/pkg/grpcx"
)

func InitGRPCxServer(
	comment *grpc2.CommentServiceServer,
	ecli *clientv3.Client,
	l logger.Logger,
) *grpcx.Server {
	type Config struct {
		Port      int      `yaml:"port"`
		EtcdAddrs []string `yaml:"etcdAddrs"`
		EtcdTTL   int64    `yaml:"etcdTTL"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.server", &cfg)
	if err != nil {
		panic(err)
	}
	server := grpc.NewServer()
	comment.Register(server)
	return &grpcx.Server{
		Server:     server,
		Port:       cfg.Port,
		Name:       "comment",
		L:          l,
		EtcdTTL:    cfg.EtcdTTL,
		EtcdClient: ecli,
		EtcdAddrs:  cfg.EtcdAddrs,
	}
}
