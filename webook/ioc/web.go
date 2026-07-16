package ioc

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	ginxlimit "webook/webook/internal/pkg/ginx/middleware/ratelimit"
	"webook/webook/internal/pkg/ratelimit"

	"webook/webook/internal/web"
	"webook/webook/internal/web/middleware"
)

func InitGin(mdl []gin.HandlerFunc, hdl *web.UserHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdl...)
	hdl.RegisterRoutes(server)
	return server
}

func InitMiddleWare(limit ratelimit.Limiter) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		cors.New(cors.Config{
			AllowOrigins:     []string{"http://localhost:3000"},
			AllowHeaders:     []string{"Content-Type", "Authorization"},
			AllowCredentials: true,
			ExposeHeaders:    []string{"x-jwt-token"},
		}),
		middleware.NewLoginMiddlewareBuilder().Build(),
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
