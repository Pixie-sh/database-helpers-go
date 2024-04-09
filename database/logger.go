package database

import (
	"context"
	"github.com/pixie-sh/logger-go/logger"
	logger2 "gorm.io/gorm/logger"
	"time"
)

type log struct {
	plog logger.Interface
}

func (l log) Printf(log string, args ...any) {
	l.plog.Log(log, args...)
}

func (l log) LogMode(_ logger2.LogLevel) logger2.Interface {
	return l
}

func (l log) Info(ctx context.Context, format string, args ...interface{}) {
	l.plog.With("ctx", ctx).Log(format, args...)
}
func (l log) Warn(ctx context.Context, format string, args ...interface{}) {
	l.plog.With("ctx", ctx).Warn(format, args...)
}
func (l log) Error(ctx context.Context, format string, args ...interface{}) {
	l.plog.With("ctx", ctx).Error(format, args...)
}
func (l log) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elap := float64(time.Since(begin).Nanoseconds()) / 1e6
	sql, rows := fc()
	l.plog.With("ctx", ctx).With("rows", rows).With("sql", sql).With("elapsed", elap).Debug("trace: %s", sql)
}
