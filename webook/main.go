package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	initPrometheus()
	app := startup.InitWebServer()
	// Consumer 在我设计下，类似于 Web，或者 GRPC 之类的，是一个顶级入口
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

func initPrometheus() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":8081", nil)
	}()
}

// initViperV1 通过命令行参数处理
