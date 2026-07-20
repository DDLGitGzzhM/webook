package ioc

import (
	"context"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"webook/webook/internal/pkg/ginx/middleware/logger"
	ginxlimit "webook/webook/internal/pkg/ginx/middleware/ratelimit"
	pkgLog "webook/webook/internal/pkg/logger"
	"webook/webook/internal/pkg/ratelimit"
	"webook/webook/internal/web"
	jwtHandler "webook/webook/internal/web/jwt"
	"webook/webook/internal/web/middleware"
)

func InitGin(mdl []gin.HandlerFunc, hdl *web.UserHandler, oauth *web.OAuth2WechatHandler, article *web.ArticleHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdl...)
	hdl.RegisterRoutes(server)
	oauth.RegisterRoutes(server)
	article.RegisterRoutes(server)
	return server
}

func InitMiddleWare(limit ratelimit.Limiter, jwt jwtHandler.Handler, log pkgLog.Logger) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		logger.NewMiddleWareBuilder(func(ctx context.Context, al *logger.Accesslog) {
			log.Info("scall", pkgLog.Field{
				Key:   "al",
				Value: al,
			})
		}).Build(),
		cors.New(cors.Config{
			AllowOrigins:     []string{"http://localhost:3000", "http://localhost:3001"},
			AllowHeaders:     []string{"Content-Type", "Authorization"},
			AllowCredentials: true,
			ExposeHeaders:    []string{"x-jwt-token", "x-refresh-token"},
		}),
		middleware.NewLoginMiddlewareBuilder(jwt).Build(),
		ginxlimit.NewBuilder(limit).Build(),
	}
}

// LimitParam 限流参数
type LimitParam struct {
	Interval time.Duration
	Rate     int
}

// InitLimitParam 初始化限流参数
func InitLimitParam() LimitParam {
	return LimitParam{
		Interval: time.Second * 100,
		Rate:     10,
	}
}
