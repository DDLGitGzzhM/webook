package ioc

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	intrv1 "webook/webook/api/proto/gen/intr/v1"
	"webook/webook/internal/service"
	"webook/webook/internal/web/client"
)

func InitIntrGRPCClient(svc service.InteractiveService) intrv1.InteractiveServiceClient {
	type Config struct {
		Addr      string
		Secure    bool
		Threshold int32
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.client.intr", &cfg)
	if err != nil {
		panic(err)
	}
	var opts []grpc.DialOption
	if cfg.Secure {
		// 上面，要去加载你的证书之类的东西
		// 启用 HTTPS
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	cc, err := grpc.Dial(cfg.Addr, opts...)
	if err != nil {
		panic(err)
	}
	remote := intrv1.NewInteractiveServiceClient(cc)
	local := client.NewInteractiveServiceAdapter(svc)
	res := client.NewGreyScaleInteractiveServiceClient(remote, local)
	res.UpdateThreshold(cfg.Threshold)
	// 我的习惯是在这里监听
	viper.OnConfigChange(func(in fsnotify.Event) {
		var cfg Config
		err = viper.UnmarshalKey("grpc.client.intr", &cfg)
		if err != nil {
			// 你可以输出日志
			return
		}
		res.UpdateThreshold(cfg.Threshold)
	})
	return res
}
