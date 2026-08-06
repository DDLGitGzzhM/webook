package startup

import (
	"webook/webook/internal/pkg/logger"
)

// InitLog 测试用空日志。
func InitLog() logger.Logger {
	return logger.NewNoOpLogger()
}
