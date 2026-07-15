package main

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	redisv9 "github.com/redis/go-redis/v9"
	mysql "gorm.io/driver/mysql"
	"gorm.io/gorm"

	"webook/webook/config/config"
	"webook/webook/internal/pkg/ginx/middleware/ratelimit"
	"webook/webook/internal/repository"
	"webook/webook/internal/repository/cache"
	smsTest "webook/webook/internal/service/sms/test"

	"webook/webook/internal/repository/dao"
	"webook/webook/internal/service"
	"webook/webook/internal/web"
	"webook/webook/internal/web/middleware"
)

func main() {
	db := initDB()
	rbd := initRedis()
	u := initUser(db, rbd)
	server := initWeb()
	u.RegisterRoutes(server)
	//server := gin.Default()
	//server.GET("/hello", func(ctx *gin.Context) {
	//	ctx.JSON(200, "你好 k8s")
	//})
	server.Run(":8080")
}

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open(config.Config.Db.DSN))
	if err != nil {
		panic(err)
	}

	if err = dao.InitTable(db); err != nil {
		panic(err)
	}
	return db
}
func initRedis() redisv9.Cmdable {
	return redisv9.NewClient(&redisv9.Options{
		Addr: config.Config.Redis.Addr,
	})
}

func initUser(db *gorm.DB, rdb redisv9.Cmdable) *web.UserHandler {
	userCache := cache.NewUserCache(rdb)
	ud := dao.NewUserDAO(db)
	repo := repository.NewUserRepository(ud, userCache)
	svc := service.NewUserService(repo)
	u := web.NewUserHandler(svc, service.NewCodeService(repository.NewCodeRepository(cache.NewCodeCache(rdb)), smsTest.NewService()))
	return u
}

func initWeb() *gin.Engine {
	server := gin.Default()
	server.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		ExposeHeaders:    []string{"x-jwt-token"},
	}))

	store, err := redis.NewStore(16, "tcp", config.Config.Redis.Addr,
		"", "", []byte("7u4hhBQpHdT0Mq2R"), []byte("j6yMxCN73DDpjDdp"))
	if err != nil {
		panic(err)
	}
	redisClint := redisv9.NewClient(&redisv9.Options{
		Addr: config.Config.Redis.Addr,
	})
	server.Use(ratelimit.NewBuilder(redisClint, 10*time.Second, 100).Build())
	server.Use(sessions.Sessions("session", store))
	server.Use(middleware.NewLoginMiddlewareBuilder().Build())

	return server
}
