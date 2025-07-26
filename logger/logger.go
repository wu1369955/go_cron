package logger

import (
	"log/slog"
)

type CronLogger struct {
	logger *slog.Logger
}

func NewCronLogger(logger *slog.Logger) *CronLogger {
	return &CronLogger{
		logger: logger,
	}
}

func (l *CronLogger) Info(msg string, keyValues ...any) {
	// 不需要详细的info日志，降为debug等级
	l.logger.With(keyValues...).Debug(msg)
}

func (l *CronLogger) Error(err error, msg string, keyValues ...any) {
	l.logger.With("err", err).With(keyValues...).Error(msg)
}
