package log

import (
	"time"

	"go.uber.org/zap"
)

var logger = newDefaultProductionLog()

type Logger = zap.Logger

func GlobalLogger() *Logger {
	return &logger
}

func SetGlobalLogger(l *Logger) {
	logger = *l
}

func Info(msg string, fields ...zap.Field) {
	logger.Info(msg, fields...)
}

func InfoE(msg string, err error, fields ...zap.Field) {
	fields = append(fields, zap.Error(err))
	logger.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	logger.Warn(msg, fields...)
}

func WarnE(msg string, err error, fields ...zap.Field) {
	fields = append(fields, zap.Error(err))
	logger.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	logger.Error(msg, fields...)
}

func ErrorE(msg string, err error, fields ...zap.Field) {
	fields = append(fields, zap.Error(err))
	logger.Error(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	logger.Debug(msg, fields...)
}

func DebugE(msg string, err error, fields ...zap.Field) {
	fields = append(fields, zap.Error(err))
	logger.Debug(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	logger.Fatal(msg, fields...)
}

func FatalE(msg string, err error, fields ...zap.Field) {
	fields = append(fields, zap.Error(err))
	logger.Fatal(msg, fields...)
}

func Panic(msg string, fields ...zap.Field) {
	logger.Panic(msg, fields...)
}

func PanicE(msg string, err error, fields ...zap.Field) {
	fields = append(fields, zap.Error(err))
	logger.Panic(msg, fields...)
}

func With(fields ...zap.Field) *Logger {
	return logger.With(fields...)
}

func StringField(key string, val string) zap.Field {
	return zap.String(key, val)
}

func AnyField(key string, val any) zap.Field {
	return zap.Any(key, val)
}

func ErrorField(err error) zap.Field {
	return zap.Error(err)
}

func IntField(key string, val int) zap.Field {
	return zap.Int(key, val)
}

func DurationField(key string, val time.Duration) zap.Field {
	return zap.Duration(key, val)
}
