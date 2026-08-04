package logger

type NopLogger struct {
}

// NewNoOpLogger 返回空实现 Logger，便于测试注入。
func NewNoOpLogger() Logger {
	return NopLogger{}
}

func (n NopLogger) Error(msg string, args ...Field) {
}

func (n NopLogger) Debug(msg string, args ...Field) {
}

func (n NopLogger) Info(msg string, args ...Field) {
}

func (n NopLogger) Warn(msg string, args ...Field) {
}
