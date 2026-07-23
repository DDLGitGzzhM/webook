package main

import (
	_ "github.com/spf13/viper/remote"

	"webook/webook/startup"
)

func main() {
	//db := initDB()
	//rbd := initRedis()
	//u := initUser(db, rbd)
	//u.RegisterRoutes(server)
	startup.InitViperDev()
	//initViperRemote()
	app := startup.InitWebServer()
	for _, c := range app.Consumers() {
		err := c.Start()
		if err != nil {
			panic(err)
		}
	}
	server := app.Web()
	//server := gin.Default()
	//server.GET("/hello", func(ctx *gin.Context) {
	//	ctx.JSON(200, "你好 k8s")
	//})
	server.Run(":8080")
}

// initViperV1 通过命令行参数处理
