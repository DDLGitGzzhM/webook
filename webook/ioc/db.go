package ioc

import (
	"time"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
	"gorm.io/plugin/prometheus"

	"webook/webook/internal/pkg/logger"
	"webook/webook/internal/repository/dao"
)

func InitDB(logger logger.Logger) *gorm.DB {
	type Config struct {
		DSN string `yaml:"dsn"`
	}
	var cfg Config
	err := viper.UnmarshalKey("db.mysql", &cfg)
	if err != nil {
		return nil
	}
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
		Logger: glogger.New(gormLoggerFunc(logger.Debug), glogger.Config{
			// 慢查询阈值 50ms ,100ms
			SlowThreshold:             time.Millisecond * 50,
			IgnoreRecordNotFoundError: true,
			LogLevel:                  glogger.Info,
		}),
	})
	if err != nil {
		panic(err)
	}

	err = db.Use(prometheus.New(prometheus.Config{
		DBName:          "webook",
		RefreshInterval: 15,
		StartServer:     false,
		MetricsCollector: []prometheus.MetricsCollector{
			&prometheus.MySQL{
				VariableNames: []string{"thread_running"},
			},
		},
	}))
	if err != nil {
		panic(err)
	}

	if err = dao.InitTable(db); err != nil {
		panic(err)
	}
	return db
}

type gormLoggerFunc func(msg string, args ...logger.Field)

func (g gormLoggerFunc) Printf(msg string, args ...interface{}) {
	g(msg, logger.Field{Key: "args", Value: args})
}
