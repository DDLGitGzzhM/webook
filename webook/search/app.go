package main

import (
	"webook/webook/pkg/grpcx"
	"webook/webook/search/events"
)

// App 控制 search 服务生命周期。
type App struct {
	server    *grpcx.Server
	consumers []events.Consumer
}
