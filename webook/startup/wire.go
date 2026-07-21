//go:build wireinject

package startup

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"

	"webook/webook/internal/pkg/logger"
	"webook/webook/internal/pkg/ratelimit"
	"webook/webook/internal/repository"
	"webook/webook/internal/repository/cache"
	"webook/webook/internal/repository/dao"
	"webook/webook/internal/repository/dao/article"
	"webook/webook/internal/service"
	"webook/webook/internal/web"
	jwtHandler "webook/webook/internal/web/jwt"
	"webook/webook/ioc"
)

var thirdProvider = wire.NewSet(ioc.InitDB, ioc.InitRedis, ioc.InitLogger)

func InitWebServer() *gin.Engine {
	wire.Build(
		ioc.InitDB, ioc.InitRedis,
		dao.NewUserDAO,
		article.NewArticleGormDao,
		wire.Bind(new(dao.UserDao), new(*dao.GormUserDAO)),
		wire.Bind(new(article.ArticleDao), new(*article.ArticleGormDao)),

		cache.NewCodeCache,
		wire.Bind(new(cache.CodeCache), new(*cache.RedisCodeCache)),
		cache.NewUserCache,
		wire.Bind(new(cache.UserCache), new(*cache.RedisUserCache)),

		repository.NewUserRepository,
		repository.NewCachedArticleRepository,
		repository.NewCodeRepository,

		wire.Bind(new(repository.UserRepository), new(*repository.CacheUserRepository)),
		wire.Bind(new(repository.CodeRepository), new(*repository.CacheCodeRepository)),
		wire.Bind(new(repository.ArticleRepository), new(*repository.CachedArticleRepository)),

		service.NewUserService,
		service.NewArticleService,
		wire.Bind(new(service.IUserService), new(*service.UserService)),
		service.NewCodeService,
		wire.Bind(new(service.ICodeService), new(*service.CodeService)),
		jwtHandler.NewRedisJwt,
		wire.Bind(new(jwtHandler.Handler), new(*jwtHandler.RedisJwt)),
		web.NewUserHandler,
		web.NewArticleHandler,
		web.NewOAuth2WechatHandler,
		ioc.InitWeChatService,
		wire.Bind(new(logger.Logger), new(*logger.ZapLogger)),
		ioc.InitLogger,
		logger.NewZapLogger,

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

func InitArticleHandler() *web.ArticleHandler {
	wire.Build(
		thirdProvider,
		logger.NewZapLogger,
		wire.Bind(new(logger.Logger), new(*logger.ZapLogger)),
		service.NewArticleService,
		web.NewArticleHandler,
		repository.NewCachedArticleRepository,
		article.NewArticleGormDao,
		wire.Bind(new(repository.ArticleRepository), new(*repository.CachedArticleRepository)),
		wire.Bind(new(article.ArticleDao), new(*article.ArticleGormDao)),
	)
	return &web.ArticleHandler{}
}
