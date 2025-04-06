package interceptors

import (
	"common_library/ctxdata"
	"context"
	"log"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

func NewUnaryLoggingInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		traceID := ""
		if v, ok := ctxdata.GetTraceID(ctx); ok {
			traceID = v
		}

		clientIP := "unknown"
		if p, ok := peer.FromContext(ctx); ok {
			clientIP = p.Addr.String()
		}

		log.Printf(
			"[gRPC Request] Method: %s | Client IP: %s | Request: %v",
			info.FullMethod,
			clientIP,
			req,
		)
		logger.Info("grpc unary request",
			zap.String("method", info.FullMethod),
			zap.String("trace_id", traceID),
			zap.String("client_ip", clientIP),
			zap.Any("request", req),
		)

		resp, err := handler(ctx, req)

		duration := time.Since(start)

		fields := []zap.Field{
			zap.String("method", info.FullMethod),
			zap.Duration("duration", duration),
			zap.String("trace_id", traceID),
		}

		if err != nil {
			fields = append(fields, zap.Error(err))
			logger.Error("request failed", fields...)
		} else {
			logger.Info("request handled", fields...)
		}

		return resp, err
	}
}

func NewStreamLoggingInterceptor(logger *zap.Logger) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		start := time.Now()

		traceID := ""
		if v, ok := ctxdata.GetTraceID(ss.Context()); ok {
			traceID = v
		}

		clientIP := "unknown"
		if p, ok := peer.FromContext(ss.Context()); ok {
			clientIP = p.Addr.String()
		}

		logger.Info("grpc stream started",
			zap.String("method", info.FullMethod),
			zap.String("trace_id", traceID),
			zap.String("client_ip", clientIP),
		)

		err := handler(srv, ss)

		duration := time.Since(start)

		fields := []zap.Field{
			zap.String("method", info.FullMethod),
			zap.Duration("duration", duration),
			zap.String("trace_id", traceID),
		}

		if err != nil {
			fields = append(fields, zap.Error(err))
			logger.Error("stream failed", fields...)
		} else {
			logger.Info("stream completed", fields...)
		}

		return err
	}
}
