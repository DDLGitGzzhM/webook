package main

import (
	"webook/webook/pkg/grpcx"
	"webook/webook/reward/events"
)

// App 控制 reward 服务生命周期。
type App struct {
	GRPCServer *grpcx.Server
	Consumer   *events.PaymentEventConsumer
}
