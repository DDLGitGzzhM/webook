package main

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
)

func main() {
	//db := initDB()
	//rbd := initRedis()
	//u := initUser(db, rbd)
	//u.RegisterRoutes(server)
	//initViperV1()
	initViperRemote()
	server := InitWebServer()
	//server := gin.Default()
	//server.GET("/hello", func(ctx *gin.Context) {
	//	ctx.JSON(200, "你好 k8s")
	//})
	server.Run(":8080")
}

func initViperRemote() {
	viper.SetConfigType("yaml")
	err := viper.AddRemoteProvider("etcd3",
		"127.0.0.1:12379", "/webook")
	if err != nil {
		return
	}
	err = viper.ReadRemoteConfig()
	if err != nil {
		panic(err)
	}
}

// initViperV1 通过命令行参数处理
func initViperV1() {
	cfile := pflag.String("config", "webook/config/config/dev.yaml", "指定文件路径")
	pflag.Parse()
	viper.SetConfigFile(*cfile)
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
}

func initViper() {
	viper.SetConfigFile("webook/config/config/dev.yaml")
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
}
