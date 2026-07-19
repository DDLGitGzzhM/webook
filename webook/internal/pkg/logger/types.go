package logger

type Logger interface {
	Error(msg string, args ...Field)
	Debug(msg string, args ...Field)
	Info(msg string, args ...Field)
	Warn(msg string, args ...Field)
}

func LoggerExample() {
	var l Logger
	l.Info("hello world")
}

type Field struct {
	Key   string
	Value any
}
