// Package wego 是一个随便取的名字
package wego

import (
	"webook/webook/internal/events"
	"webook/webook/internal/pkg/ginx"
	"webook/webook/pkg/grpcx"
)

// App 当你在 wire 里面使用这个结构体的时候，要注意不是所有的服务都需要全部字段，
// 那么在 wire 的时候就不要使用 * 了
type App struct {
	GRPCServer *grpcx.Server
	WebServer  *ginx.Server
	Consumers  []events.Consumer
}
