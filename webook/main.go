package main

import (
	"context"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	_ "github.com/spf13/viper/remote"

	"webook/webook/ioc"
	"webook/webook/startup"
)

func main() {
	//db := initDB()
	//rbd := initRedis()
	//u := initUser(db, rbd)
	//u.RegisterRoutes(server)
	startup.InitViperDev()
	//initViperRemote()
	closeFunc := ioc.InitOTEL()
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
	// 一分钟内你要关完，要退出
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	closeFunc(ctx)
}

func initPrometheus() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":8081", nil)
	}()
}

// initViperV1 通过命令行参数处理
