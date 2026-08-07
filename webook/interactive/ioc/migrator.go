package ioc

import (
	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"gorm.io/gorm"

	ievents "webook/webook/interactive/events"
	"webook/webook/internal/pkg/ginx"
	"webook/webook/internal/pkg/logger"
	"webook/webook/internal/repository/dao"
	"webook/webook/pkg/gormx/connpool"
	"webook/webook/pkg/migrator/events"
	"webook/webook/pkg/migrator/events/fixer"
	"webook/webook/pkg/migrator/scheduler"
)

const topic = "migrator_interactives"

func InitFixDataConsumer(l logger.Logger,
	src SrcDB,
	dst DstDB,
	client sarama.Client) *fixer.Consumer[dao.Interactive] {
	res, err := fixer.NewConsumer[dao.Interactive](client, l,
		topic, src, dst)
	if err != nil {
		panic(err)
	}
	return res
}

func InitMigradatorProducer(p sarama.SyncProducer) events.Producer {
	return events.NewSaramaProducer(p, topic)
}

func InitMySQLBinlogConsumer(
	client sarama.Client,
	l logger.Logger,
	src SrcDB,
	dst DstDB,
	p events.Producer,
) *ievents.MySQLBinlogConsumer[dao.Interactive] {
	return ievents.NewMySQLBinlogConsumer[dao.Interactive](
		client, l, "interactives", (*gorm.DB)(src), (*gorm.DB)(dst), p)
}

func InitMigratorWeb(
	l logger.Logger,
	src SrcDB,
	dst DstDB,
	pool *connpool.DoubleWritePool,
	producer events.Producer,
) *ginx.Server {
	// 在这里，有多少张表，你就初始化多少个 scheduler
	intrSch := scheduler.NewScheduler[dao.Interactive](l, src, dst, pool, producer)
	engine := gin.Default()
	ginx.L = l
	ginx.InitCounter(prometheus.CounterOpts{
		Namespace: "geekbang_daming",
		Subsystem: "webook_intr_admin",
		Name:      "http_biz_code",
		Help:      "HTTP 的业务错误码",
	})
	intrSch.RegisterRoutes(engine.Group("/migrator"))
	addr := viper.GetString("migrator.web.addr")
	return &ginx.Server{
		Addr:   addr,
		Engine: engine,
	}
}
