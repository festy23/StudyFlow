package main

import (
	"common_library/logging"
	"common_library/metadata"
	"context"
	"fmt"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
	"os"
	"os/signal"
	"syscall"
	"userservice/internal/config"
	"userservice/internal/data"
	"userservice/internal/db"
	"userservice/internal/handler"
	"userservice/internal/service"
	pb "userservice/pkg/api"
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

	cfg, err := config.New()
	if err != nil {
		logger.Fatal(ctx, "cannot create config", zap.Error(err))
	}

	database, err := db.New(ctx, cfg)
	if err != nil {
		logger.Fatal(ctx, "cannot create db", zap.Error(err))
	}

	userRepo := data.NewUserRepository(database)
	tsRepo := data.NewTutorStudentRepository(database)

	userService := service.NewUserService(userRepo, tsRepo, cfg.TelegramSecret)

	userHandler := handler.NewUserServiceServer(userService)

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

	pb.RegisterUserServiceServer(server, userHandler)

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
