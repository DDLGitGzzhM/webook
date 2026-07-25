package startup

import (
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"

	"webook/webook/internal/events"
)

type App struct {
	web       *gin.Engine
	consumers []events.Consumer
	cron      *cron.Cron
}

func (a *App) Web() *gin.Engine {
	return a.web
}

func (a *App) Consumers() []events.Consumer {
	return a.consumers
}

func (a *App) Cron() *cron.Cron {
	return a.cron
}
