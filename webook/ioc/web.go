package ioc

import (
	"context"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"webook/webook/internal/pkg/ginx"
	"webook/webook/internal/pkg/ginx/middleware/logger"
	"webook/webook/internal/pkg/ginx/middleware/metric"
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
	(&web.ObservabilityHandler{}).RegisterRoutes(server)
	return server
}

func InitMiddleWare(limit ratelimit.Limiter, jwt jwtHandler.Handler, log pkgLog.Logger) []gin.HandlerFunc {
	ginx.L = log
	ginx.InitCounter(prometheus.CounterOpts{
		Namespace: "geekbang_daming",
		Subsystem: "webook",
		Name:      "http_biz_code",
		Help:      "HTTP 的业务错误码",
	})
	return []gin.HandlerFunc{
		logger.NewMiddleWareBuilder(func(ctx context.Context, al *logger.Accesslog) {
			log.Info("scall", pkgLog.Field{
				Key:   "al",
				Value: al,
			})
		}).Build(),
		(&metric.MiddlewareBuilder{
			Namespace:  "geekbang_daming",
			Subsystem:  "webook",
			Name:       "gin_http",
			Help:       "统计 GIN 的 HTTP 接口",
			InstanceID: "my-instance-1",
		}).Build(),
		otelgin.Middleware("webook"),
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
