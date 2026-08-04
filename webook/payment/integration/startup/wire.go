//go:build wireinject

package startup

import (
	"github.com/google/wire"

	"webook/webook/payment/events"
	"webook/webook/payment/ioc"
	"webook/webook/payment/repository"
	"webook/webook/payment/repository/dao"
	"webook/webook/payment/service/wechat"
)

var thirdPartySet = wire.NewSet(ioc.InitLogger, InitTestDB)

func InitNopProducer() events.Producer {
	return events.NopProducer{}
}

var wechatNativeSvcSet = wire.NewSet(
	ioc.InitWechatClient,
	dao.NewPaymentGORMDAO,
	repository.NewPaymentRepository,
	InitNopProducer,
	ioc.InitWechatNativeService,
	ioc.InitWechatConfig)

func InitWechatNativeService() *wechat.NativePaymentService {
	wire.Build(wechatNativeSvcSet, thirdPartySet)
	return new(wechat.NativePaymentService)
}
