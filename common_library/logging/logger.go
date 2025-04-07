package logging

import (
	"common_library/ctxdata"
	"context"
	"go.uber.org/zap"
)

type loggerKey struct{}

const (
	requestID = "request_id"
)

var (
	loggerKeyInstance = loggerKey{}
)

type Logger struct {
	l *zap.Logger
}

func New(zapLogger *zap.Logger) *Logger {
	return &Logger{zapLogger}
}

func ContextWithLogger(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, loggerKeyInstance, logger)
}

func GetFromContext(ctx context.Context) (*Logger, bool) {
	logger, ok := ctx.Value(loggerKeyInstance).(*Logger)
	return logger, ok
}

func (l *Logger) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	fields = fieldsWithTraceID(ctx, fields)
	l.l.Debug(msg, fields...)
}

func (l *Logger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	fields = fieldsWithTraceID(ctx, fields)
	l.l.Info(msg, fields...)
}

func (l *Logger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	fields = fieldsWithTraceID(ctx, fields)
	l.l.Warn(msg, fields...)
}

func (l *Logger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	fields = fieldsWithTraceID(ctx, fields)
	l.l.Error(msg, fields...)
}

func (l *Logger) Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	fields = fieldsWithTraceID(ctx, fields)
	l.l.Fatal(msg, fields...)
}

func fieldsWithTraceID(ctx context.Context, fields []zap.Field) []zap.Field {
	if traceId, ok := ctxdata.GetTraceID(ctx); ok {
		fields = append(fields, zap.String(requestID, traceId))
	}
	return fields
}
