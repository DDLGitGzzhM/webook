//go:build wireinject

package main

import (
	"github.com/google/wire"

	"webook/webook/payment/ioc"
	"webook/webook/payment/web"
)

func InitApp() *App {
	wire.Build(
		//ioc.InitWechatClient,
		//ioc.InitWechatNativeService,
		ioc.InitWechatConfig,
		ioc.InitWechatNotifyHandler,
		ioc.InitNilNativePaymentService,
		web.NewWechatHandler,
		ioc.InitGinServer,
		ioc.InitLogger,
		wire.Struct(new(App), "WebServer"))
	return new(App)
}
