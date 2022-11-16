package xlog

import (
	"context"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"
	"saas_service/pkg/setting"
)

var Logger *zap.Logger

func InitLogger(config *setting.Config) *zap.Logger {
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	infoLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		if config.ServerCfg.RunMode != "release" {
			return lvl <= zapcore.InfoLevel
		}
		return lvl <= zapcore.InfoLevel
	})

	errorLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})

	infoFilePath := config.AppCfg.LogSavePath + "/" + config.AppCfg.Name + config.AppCfg.LogFileExt
	infoWriter := getLogWriter(infoFilePath)

	errorFilePath := config.AppCfg.LogSavePath + "/" + config.AppCfg.Name + ".wf" + config.AppCfg.LogFileExt
	errorWriter := getLogWriter(errorFilePath)
	// 最后创建具体的Logger
	core := zapcore.NewTee(
		zapcore.NewCore(encoder, zapcore.AddSync(infoWriter), infoLevel),
		zapcore.NewCore(encoder, zapcore.AddSync(errorWriter), errorLevel),
	)
	logger := zap.New(core)

	zap.ReplaceGlobals(logger)
	return logger
}

func getLogWriter(path string) zapcore.WriteSyncer {
	dir, _ := filepath.Split(path)
	err := os.MkdirAll(dir, os.FileMode(0755))
	if err != nil {
		panic(err)
	}

	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.FileMode(0666))
	if err != nil {
		file.Close()
		panic(err)
	}
	return zapcore.AddSync(file)
}

func XInfo(ctx context.Context, msg string, fields ...zap.Field) {

	_requestID := zap.String("x-request-id", GetStringFromCtx(ctx, "__x_request_id"))

	fields = append(fields, _requestID)
	Logger.Info(msg, fields...)
}

func XError(ctx context.Context, msg string, fields ...zap.Field) {

	_requestID := zap.String("x-request-id", GetStringFromCtx(ctx, "__x_request_id"))

	fields = append(fields, _requestID)
	Logger.Error(msg, fields...)
}
func XErrorF(ctx context.Context, msg string, fields ...zap.Field) {

	_requestID := zap.String("x-request-id", GetStringFromCtx(ctx, "__x_request_id"))

	fields = append(fields, _requestID)
	Logger.Error(msg, fields...)
}

func XWarn(ctx context.Context, msg string, fields ...zap.Field) {

	_requestID := zap.String("x-request-id", GetStringFromCtx(ctx, "__x_request_id"))

	fields = append(fields, _requestID)
	Logger.Warn(msg, fields...)
}

func XDebug(ctx context.Context, msg string, fields ...zap.Field) {

	_requestID := zap.String("x-request-id", GetStringFromCtx(ctx, "__x_request_id"))

	fields = append(fields, _requestID)
	Logger.Debug(msg, fields...)
}

func XSInfo(ctx context.Context, args ...interface{}) {
	args = append(args, " request-id###")
	args = append(args, GetStringFromCtx(ctx, "__x_request_id"))
	Logger.Sugar().Info(args...)
}

func XSError(ctx context.Context, args ...interface{}) {
	args = append(args, " request-id###")
	args = append(args, GetStringFromCtx(ctx, "__x_request_id"))
	Logger.Sugar().Error(args...)
}

func XSErrorF(ctx context.Context, template string, args ...interface{}) {
	args = append(args, " request-id###")
	args = append(args, GetStringFromCtx(ctx, "__x_request_id"))
	Logger.Sugar().Errorf(template, args...)
}

func XSWarn(ctx context.Context, args ...interface{}) {
	args = append(args, " request-id###")
	args = append(args, GetStringFromCtx(ctx, "__x_request_id"))
	Logger.Sugar().Warn(args...)
}

func XSDebug(ctx context.Context, args ...interface{}) {
	args = append(args, " request-id###")
	args = append(args, GetStringFromCtx(ctx, "__x_request_id"))
	Logger.Sugar().Debug(args...)
}

func GetStringFromCtx(ctx context.Context, key string) string {
	var v string
	val := ctx.Value(key)
	if val != nil {
		v = val.(string)
	}
	return v
}
