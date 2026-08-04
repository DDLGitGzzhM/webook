package main

import "webook/webook/pkg/grpcx"

// App 控制 account 服务生命周期。
type App struct {
	GRPCServer *grpcx.Server
}
