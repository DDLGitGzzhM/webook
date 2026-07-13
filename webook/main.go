package main

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	mysql "gorm.io/driver/mysql"
	"gorm.io/gorm"

	redisv9 "github.com/redis/go-redis/v9"
	"webook/webook/internal/pkg/ginx/middleware/ratelimit"
	"webook/webook/internal/repository"
	"webook/webook/internal/repository/dao"
	"webook/webook/internal/service"
	"webook/webook/internal/web"
	"webook/webook/internal/web/middleware"
)

func main() {
	db := initDB()
	u := initUser(db)
	server := initWeb()
	u.RegisterRoutes(server)
	server.Run(":8080")
}

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:root@tcp(127.0.0.1:13316)/webook"))
	if err != nil {
		panic(err)
	}

	if err = dao.InitTable(db); err != nil {
		panic(err)
	}
	return db
}

func initUser(db *gorm.DB) *web.UserHandler {
	ud := dao.NewUserDAO(db)
	repo := repository.NewUserRepository(ud)
	svc := service.NewUserService(repo)
	u := web.NewUserHandler(svc)
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

	store, err := redis.NewStore(16, "tcp", "localhost:6380",
		"", "", []byte("7u4hhBQpHdT0Mq2R"), []byte("j6yMxCN73DDpjDdp"))
	if err != nil {
		panic(err)
	}
	redisClint := redisv9.NewClient(&redisv9.Options{
		Addr: "localhost:6380",
	})
	server.Use(ratelimit.NewBuilder(redisClint, 10*time.Second, 100).Build())
	server.Use(sessions.Sessions("session", store))
	server.Use(middleware.NewLoginMiddlewareBuilder().Build())

	return server
}
