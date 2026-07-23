package startup

import (
	"github.com/gin-gonic/gin"

	"webook/webook/internal/events"
)

type App struct {
	web       *gin.Engine
	consumers []events.Consumer
}

func (a *App) Web() *gin.Engine {
	return a.web
}

func (a *App) Consumers() []events.Consumer {
	return a.consumers
}
