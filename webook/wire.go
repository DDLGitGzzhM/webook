//go:build wireinject

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"

	"webook/webook/internal/pkg/ratelimit"
	"webook/webook/internal/repository"
	"webook/webook/internal/repository/cache"
	"webook/webook/internal/repository/dao"
	"webook/webook/internal/service"
	"webook/webook/internal/web"
	"webook/webook/ioc"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		ioc.InitDB, ioc.InitRedis,
		dao.NewUserDAO,
		wire.Bind(new(dao.UserDao), new(*dao.GormUserDAO)),

		cache.NewCodeCache,
		wire.Bind(new(cache.CodeCache), new(*cache.RedisCodeCache)),
		cache.NewUserCache,
		wire.Bind(new(cache.UserCache), new(*cache.RedisUserCache)),

		repository.NewUserRepository,
		wire.Bind(new(repository.UserRepository), new(*repository.CacheUserRepository)),
		repository.NewCodeRepository,
		wire.Bind(new(repository.CodeRepository), new(*repository.CacheCodeRepository)),

		service.NewUserService,
		wire.Bind(new(service.IUserService), new(*service.UserService)),
		service.NewCodeService,
		wire.Bind(new(service.ICodeService), new(*service.CodeService)),
		web.NewUserHandler,
		web.NewJWTHandler,
		web.NewOAuth2WechatHandler,
		ioc.InitWeChatService,

		ratelimit.NewRedisSlideWindowLimiter,
		wire.Bind(new(ratelimit.Limiter), new(*ratelimit.RedisSlideWindowLimiter)),

		ioc.InitGin,
		ioc.InitMiddleWare,
		ioc.InitSMSService,
		ioc.InitLimitParam,
		wire.FieldsOf(new(ioc.LimitParam), "Interval", "Rate"),
	)
	return new(gin.Engine)
}
