package main

import "webook/webook/internal/pkg/ginx"

// App 控制 payment 服务生命周期。
type App struct {
	WebServer *ginx.Server
}
