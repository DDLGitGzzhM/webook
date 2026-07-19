package logger

import (
	"github.com/samber/lo"
	"go.uber.org/zap"
)

type ZapLogger struct {
	logger *zap.Logger
}

func NewZapLogger(l *zap.Logger) *ZapLogger {
	return &ZapLogger{
		logger: l,
	}
}

func (z *ZapLogger) Error(msg string, args ...Field) {
	z.logger.Error(msg, z.toZapField(args...)...)
}

func (z *ZapLogger) Debug(msg string, args ...Field) {
	z.logger.Debug(msg, z.toZapField(args...)...)
}

func (z *ZapLogger) Info(msg string, args ...Field) {
	z.logger.Info(msg, z.toZapField(args...)...)
}

func (z *ZapLogger) Warn(msg string, args ...Field) {
	z.logger.Warn(msg, z.toZapField(args...)...)
}

func (z *ZapLogger) toZapField(args ...Field) []zap.Field {
	return lo.Map(args, func(item Field, _ int) zap.Field {
		return zap.Any(item.Key, item.Value)
	})
}
