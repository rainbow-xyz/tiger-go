package xlog

import (
	"context"
	"errors"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type ZGLogger struct {
	ZapLogger                 *zap.Logger
	LogLevel                  gormlogger.LogLevel
	SlowThreshold             time.Duration
	SkipCallerLookup          bool
	IgnoreRecordNotFoundError bool
}

func NewZGLogger(zapLogger *zap.Logger) ZGLogger {
	return ZGLogger{
		ZapLogger:                 zapLogger,
		LogLevel:                  gormlogger.Warn,
		SlowThreshold:             100 * time.Millisecond,
		SkipCallerLookup:          false,
		IgnoreRecordNotFoundError: false,
	}
}

func (l ZGLogger) SetAsDefault() {
	gormlogger.Default = l
}

func (l ZGLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	return ZGLogger{
		ZapLogger:                 l.ZapLogger,
		SlowThreshold:             l.SlowThreshold,
		LogLevel:                  level,
		SkipCallerLookup:          l.SkipCallerLookup,
		IgnoreRecordNotFoundError: l.IgnoreRecordNotFoundError,
	}
}

func (l ZGLogger) Info(ctx context.Context, str string, args ...interface{}) {
	if l.LogLevel < gormlogger.Info {
		return
	}
	s := str + " request-id: "
	args = append(args, " request-id: ")
	args = append(args, GetStringFromCtx(ctx, "__x_request_id"))
	l.logger().Sugar().Infof(s, args...)
}

func (l ZGLogger) Warn(ctx context.Context, str string, args ...interface{}) {
	if l.LogLevel < gormlogger.Warn {
		return
	}
	s := str + " request-id: "
	args = append(args, " request-id: ")
	args = append(args, GetStringFromCtx(ctx, "__x_request_id"))
	l.logger().Sugar().Warnf(s, args...)
}

func (l ZGLogger) Error(ctx context.Context, str string, args ...interface{}) {
	if l.LogLevel < gormlogger.Error {
		return
	}
	s := str + " request-id: "
	args = append(args, " request-id: ")
	args = append(args, GetStringFromCtx(ctx, "__x_request_id"))
	l.logger().Sugar().Errorf(s, args...)
}

func (l ZGLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	requestID := GetStringFromCtx(ctx, "__x_request_id")
	if l.LogLevel <= 0 {
		return
	}
	elapsed := time.Since(begin)
	switch {
	case err != nil && l.LogLevel >= gormlogger.Error && (!l.IgnoreRecordNotFoundError || !errors.Is(err, gorm.ErrRecordNotFound)):
		sql, rows := fc()
		l.logger().Error("trace", zap.String("x-request-id", requestID), zap.Duration("elapsed", elapsed), zap.Int64("rows", rows), zap.String("sql", sql))
	case l.SlowThreshold != 0 && elapsed > l.SlowThreshold && l.LogLevel >= gormlogger.Warn:
		sql, rows := fc()
		l.logger().Warn("trace", zap.String("x-request-id", requestID), zap.Duration("elapsed", elapsed), zap.Int64("rows", rows), zap.String("sql", sql))
	case l.LogLevel >= gormlogger.Info:
		sql, rows := fc()
		l.logger().Info("trace", zap.String("x-request-id", requestID), zap.Duration("elapsed", elapsed), zap.Int64("rows", rows), zap.String("sql", sql))
	}
}

var (
	gormPackage    = filepath.Join("gorm.io", "gorm")
	zapgormPackage = filepath.Join("saas_service/pkg", "xlog")
)

func (l ZGLogger) logger() *zap.Logger {
	for i := 2; i < 15; i++ {
		_, file, _, ok := runtime.Caller(i)
		switch {
		case !ok:
		case strings.HasSuffix(file, "_test.go"):
		case strings.Contains(file, gormPackage):
		case strings.Contains(file, zapgormPackage):
		default:
			return l.ZapLogger.WithOptions(zap.AddCaller(), zap.AddCallerSkip(i))
		}
	}
	return l.ZapLogger
}
