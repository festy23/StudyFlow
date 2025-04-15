package logger

import (
	"go.uber.org/zap"
)

type Logger struct {
	ZapLogger *zap.Logger
}

func New() *Logger {
	zapLogger, _ := zap.NewProduction()
	return &Logger{ZapLogger: zapLogger}
}

func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.ZapLogger.Info(msg, fields...)
}

func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.ZapLogger.Error(msg, fields...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.ZapLogger.Sugar().Errorf(format, args...)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.ZapLogger.Sugar().Infof(format, args...)
}

func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.ZapLogger.Sugar().Fatalf(format, args...)
}

func (l *Logger) Sync() error {
	return l.ZapLogger.Sync()
}
