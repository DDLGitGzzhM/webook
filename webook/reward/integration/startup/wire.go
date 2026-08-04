//go:build wireinject

package startup

import (
	"github.com/google/wire"

	accountv1 "webook/webook/api/proto/gen/account/v1"
	pmtv1 "webook/webook/api/proto/gen/payment/v1"
	"webook/webook/reward/repository"
	"webook/webook/reward/repository/cache"
	"webook/webook/reward/repository/dao"
	"webook/webook/reward/service"
)

var thirdPartySet = wire.NewSet(InitTestDB, InitLogger, InitRedis)

func InitNilAccountClient() accountv1.AccountServiceClient {
	return nil
}

func InitWechatNativeSvc(
	client pmtv1.WechatPaymentServiceClient,
) *service.WechatNativeRewardService {
	wire.Build(service.NewWechatNativeRewardService,
		thirdPartySet,
		InitNilAccountClient,
		cache.NewRewardRedisCache,
		repository.NewRewardRepository, dao.NewRewardGORMDAO)
	return new(service.WechatNativeRewardService)
}
