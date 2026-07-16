package ioc

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"webook/webook/internal/pkg/ginx/middleware/ratelimit"
	"webook/webook/internal/web"
	"webook/webook/internal/web/middleware"
)

func InitGin(mdl []gin.HandlerFunc, hdl *web.UserHandler) *gin.Engine {
	server := gin.Default()
	hdl.RegisterRoutes(server)
	return server
}

func InitMiddleWare(redisClint redis.Cmdable) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		cors.New(cors.Config{
			AllowOrigins:     []string{"http://localhost:3000"},
			AllowHeaders:     []string{"Content-Type", "Authorization"},
			AllowCredentials: true,
			ExposeHeaders:    []string{"x-jwt-token"},
		}),
		ratelimit.NewBuilder(redisClint, 10*time.Second, 100).Build(),
		middleware.NewLoginMiddlewareBuilder().Build(),
	}
}
