package logger

import (
	"context"
	"go.uber.org/zap"
)

type Logger struct {
	ZapLogger *zap.Logger
}

func New() *Logger {
	zapLogger, _ := zap.NewDevelopment()
	return &Logger{ZapLogger: zapLogger}
}

func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.ZapLogger.Info(msg, fields...)
}

func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.ZapLogger.Warn(msg, fields...)
}

func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.ZapLogger.Error(msg, fields...)
}

func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.ZapLogger.Debug(msg, fields...)
}

func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.ZapLogger.Fatal(msg, fields...)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.ZapLogger.Sugar().Infof(format, args...)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.ZapLogger.Sugar().Warnf(format, args...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.ZapLogger.Sugar().Errorf(format, args...)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.ZapLogger.Sugar().Debugf(format, args...)
}

func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.ZapLogger.Sugar().Fatalf(format, args...)
}

func (l *Logger) InfoContext(ctx context.Context, msg string, fields ...zap.Field) {
	l.ZapLogger.Info(msg, append(fields, zap.Any("context", ctx))...)
}

func (l *Logger) WarnContext(ctx context.Context, msg string, fields ...zap.Field) {
	l.ZapLogger.Warn(msg, append(fields, zap.Any("context", ctx))...)
}

func (l *Logger) ErrorContext(ctx context.Context, msg string, fields ...zap.Field) {
	l.ZapLogger.Error(msg, append(fields, zap.Any("context", ctx))...)
}

func (l *Logger) DebugContext(ctx context.Context, msg string, fields ...zap.Field) {
	l.ZapLogger.Debug(msg, append(fields, zap.Any("context", ctx))...)
}

func (l *Logger) Sync() error {
	return l.ZapLogger.Sync()
}

func (l *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{ZapLogger: l.ZapLogger.With(fields...)}
}
