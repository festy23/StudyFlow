package main

import (
	"common_library/logging"
	"common_library/metadata"
	"context"
	"fileservice/internal/config"
	"fileservice/internal/data"
	"fileservice/internal/db"
	"fileservice/internal/handler"
	"fileservice/internal/s3_client"
	"fileservice/internal/service"
	pb "fileservice/pkg/api"
	"fmt"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx := context.Background()
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	logger := logging.New(zapLogger)

	ctx = logging.ContextWithLogger(ctx, logger)

	cfg, err := config.New()
	if err != nil {
		logger.Fatal(ctx, "cannot create config", zap.Error(err))
	}

	database, err := db.New(ctx, cfg)
	if err != nil {
		logger.Fatal(ctx, "cannot create db", zap.Error(err))
	}

	repo := data.NewFileRepository(database)

	s3Client, err := s3_client.New(ctx, cfg)
	if err != nil {
		logger.Fatal(ctx, "cannot create S3 client", zap.Error(err))
	}

	fileService, err := service.NewFileService(ctx, repo, s3Client, "user-files", cfg.GatewayPublicUrl, cfg.S3Endpoint)
	if err != nil {
		logger.Fatal(ctx, "cannot create fileService", zap.Error(err))
	}

	fileHandler := handler.NewFileHandler(fileService)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		logger.Fatal(ctx, "cannot create listener", zap.Error(err))
	}

	server := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			metadata.NewMetadataUnaryInterceptor(),
			logging.NewUnaryLoggingInterceptor(logger),
		)),
	)

	pb.RegisterFileServiceServer(server, fileHandler)

	logger.Info(ctx, "Starting gRPC server...", zap.Int("port", cfg.GRPCPort))
	go func() {
		if err := server.Serve(listener); err != nil {
			logger.Fatal(ctx, "failed to serve", zap.Error(err))
		}
	}()

	select {
	case <-ctx.Done():
		server.Stop()
		logger.Info(ctx, "Server Stopped")
	}
}
