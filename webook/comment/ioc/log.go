package ioc

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"webook/webook/internal/pkg/logger"
)

func InitLogger() logger.Logger {
	// 这里我们用一个小技巧，
	// 就是直接使用 zap 本身的配置结构体来处理
	//cfg := zap.NewDevelopmentConfig()

	//cfg.OutputPaths = []string{"./logs/app.log"}
	//err := viper.UnmarshalKey("log", &cfg)
	//if err != nil {
	//	panic(err)
	//}

	lumberLogger := &lumberjack.Logger{
		// 要注意，得有权限
		Filename:   "/var/log/comment.log",
		MaxSize:    50,
		MaxBackups: 3,
		MaxAge:     7,
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewDevelopmentEncoderConfig()),
		zapcore.AddSync(lumberLogger),
		zapcore.DebugLevel,
	)

	l := zap.New(core)

	return logger.NewZapLogger(l)
}
