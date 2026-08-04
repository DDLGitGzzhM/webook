package main

import (
	"webook/webook/internal/pkg/ginx"
	"webook/webook/pkg/grpcx"
)

// App 控制 payment 服务生命周期。
type App struct {
	WebServer  *ginx.Server
	GRPCServer *grpcx.Server
}
