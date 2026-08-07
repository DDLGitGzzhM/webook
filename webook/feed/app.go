package main

import (
	"webook/webook/feed/events"
	"webook/webook/pkg/grpcx"
)

// App 控制 feed 服务生命周期。
type App struct {
	server    *grpcx.Server
	consumers []events.Consumer
}
