package main

import (
	"fmt"
	"log"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	initViperV1()
	// 搞成依赖注入
	app := InitAPP()
	for _, c := range app.consumers {
		err := c.Start()
		if err != nil {
			panic(err)
		}
	}
	go func() {
		err := app.webAdmin.Start()
		log.Println(err)
	}()
	err := app.server.Serve()
	log.Println(err)
}

func initViperV1() {
	cfile := pflag.String("config",
		"config/dev.yaml", "指定配置文件路径")
	pflag.Parse()
	viper.SetConfigFile(*cfile)
	// 实时监听配置变更
	viper.WatchConfig()
	// 只能告诉你文件变了，不能告诉你，文件的哪些内容变了
	viper.OnConfigChange(func(in fsnotify.Event) {
		fmt.Println(in.Name, in.Op)
		fmt.Println(viper.GetString("db.dsn"))
	})
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}
