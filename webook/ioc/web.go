package ioc

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	ginxlimit "webook/webook/internal/pkg/ginx/middleware/ratelimit"
	"webook/webook/internal/pkg/ratelimit"
	jwtHandler "webook/webook/internal/web/jwt"

	"webook/webook/internal/web"
	"webook/webook/internal/web/middleware"
)

func InitGin(mdl []gin.HandlerFunc, hdl *web.UserHandler, oauth *web.OAuth2WechatHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdl...)
	hdl.RegisterRoutes(server)
	oauth.RegisterRoutes(server)
	return server
}

func InitMiddleWare(limit ratelimit.Limiter, jwt jwtHandler.Handler) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		cors.New(cors.Config{
			AllowOrigins:     []string{"http://localhost:3000"},
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
