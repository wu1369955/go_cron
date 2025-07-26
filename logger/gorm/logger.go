package gorm

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime"
	"time"

	gormlib "gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type Logger struct {
	logger                    *slog.Logger
	LogLevel                  gormlogger.LogLevel
	SlowThreshold             time.Duration
	IgnoreRecordNotFoundError bool
	// skip caller 要跳过的caller层级数量
	// https://pkg.go.dev/log/slog#hdr-Wrapping_output_methods
	// https://pkg.go.dev/log/slog#example-package-Wrapping
	// 默认为4
	CallerSkip int
}

func New(logger *slog.Logger) Logger {
	return Logger{
		logger:                    logger,
		LogLevel:                  gormlogger.Info,
		SlowThreshold:             100 * time.Millisecond,
		IgnoreRecordNotFoundError: false,
		CallerSkip:                4,
	}
}

func (l Logger) SetAsDefault() {
	gormlogger.Default = l
}

func (l Logger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newlogger := l
	newlogger.LogLevel = level
	return &newlogger
}

func (l Logger) Info(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= gormlogger.Info {
		// 需要手动设置source
		// https://pkg.go.dev/log/slog#hdr-Wrapping_output_methods
		// https://pkg.go.dev/log/slog#example-package-Wrapping
		var pcs [1]uintptr
		runtime.Callers(l.CallerSkip, pcs[:])
		r := slog.NewRecord(time.Now(), slog.LevelInfo, fmt.Sprintf(msg, data...), pcs[0])
		_ = l.logger.Handler().Handle(ctx, r)
	}
}

func (l Logger) Warn(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= gormlogger.Warn {
		var pcs [1]uintptr
		runtime.Callers(l.CallerSkip, pcs[:])
		r := slog.NewRecord(time.Now(), slog.LevelWarn, fmt.Sprintf(msg, data...), pcs[0])
		_ = l.logger.Handler().Handle(ctx, r)
	}
}

func (l Logger) Error(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= gormlogger.Error {
		var pcs [1]uintptr
		runtime.Callers(l.CallerSkip, pcs[:])
		r := slog.NewRecord(time.Now(), slog.LevelError, fmt.Sprintf(msg, data...), pcs[0])
		_ = l.logger.Handler().Handle(ctx, r)
	}
}

func (l Logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	attrs := []slog.Attr{
		slog.String("duration", fmt.Sprintf("%.3fms", float64(elapsed.Nanoseconds())/1e6)),
		slog.String("sql", sql),
	}
	if rows >= 0 {
		attrs = append(attrs, slog.Int64("rows", rows))
	}
	if err != nil {
		attrs = append(attrs, slog.Any("error", err))
	}

	var pcs [1]uintptr
	runtime.Callers(l.CallerSkip, pcs[:])

	switch {
	case err != nil && l.LogLevel >= gormlogger.Error && (!l.IgnoreRecordNotFoundError || !errors.Is(err, gormlib.ErrRecordNotFound)):
		r := slog.NewRecord(time.Now(), slog.LevelError, "", pcs[0])
		r.AddAttrs(attrs...)
		_ = l.logger.Handler().Handle(ctx, r)
	case l.SlowThreshold != 0 && elapsed > l.SlowThreshold && l.LogLevel >= gormlogger.Warn:
		r := slog.NewRecord(time.Now(), slog.LevelWarn, "", pcs[0])
		r.AddAttrs(attrs...)
		_ = l.logger.Handler().Handle(ctx, r)
	case l.LogLevel >= gormlogger.Info:
		r := slog.NewRecord(time.Now(), slog.LevelInfo, "", pcs[0])
		r.AddAttrs(attrs...)
		_ = l.logger.Handler().Handle(ctx, r)
	}
}
