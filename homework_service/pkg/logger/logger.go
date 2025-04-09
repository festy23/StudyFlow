package logger

import (
	"go.uber.org/zap"
)

type Logger struct {
	logger *zap.Logger
}

func New() *Logger {
	zapLogger, _ := zap.NewProduction()
	return &Logger{logger: zapLogger}
}

func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.logger.Info(msg, fields...)
}

func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.logger.Error(msg, fields...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.logger.Sugar().Errorf(format, args...)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.logger.Sugar().Infof(format, args...)
}

func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.logger.Sugar().Fatalf(format, args...)
}

func (l *Logger) Sync() error {
	return l.logger.Sync()
}
