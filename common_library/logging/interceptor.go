package logging

import (
	"context"
	"google.golang.org/grpc/metadata"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

func NewUnaryLoggingInterceptor(logger *Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		clientIP := "unknown"
		if p, ok := peer.FromContext(ctx); ok {
			clientIP = p.Addr.String()
		}

		fields := []zap.Field{
			zap.String("method", info.FullMethod),
			zap.String("client_ip", clientIP),
			zap.Any("request", req),
		}

		if md, ok := metadata.FromIncomingContext(ctx); ok {
			fields = append(fields, zap.Any("metadata", md))
		}

		logger.Info(ctx, "grpc unary request",
			fields...,
		)

		ctx = ContextWithLogger(ctx, logger)

		resp, err := handler(ctx, req)

		duration := time.Since(start)

		fields = []zap.Field{
			zap.String("method", info.FullMethod),
			zap.Duration("duration", duration),
		}

		if err != nil {
			fields = append(fields, zap.Error(err))
			logger.Error(ctx, "request failed", fields...)
		} else {
			logger.Info(ctx, "request handled", fields...)
		}

		return resp, err
	}
}
