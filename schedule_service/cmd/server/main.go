package main

import (
	"common_library/logging"
	"common_library/metadata"
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"schedule_service/internal/config"
	"schedule_service/internal/database/postgres"
	service "schedule_service/internal/service/service"
	pb "schedule_service/pkg/api"
	"syscall"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	zapLogger, err := zap.NewDevelopment()

	if err != nil {
		zap.Error(err)
	}

	logger := logging.New(zapLogger)

	ctx = logging.ContextWithLogger(ctx, logger)

	cfg := config.GetConfig()
	if err != nil {
		logger.Fatal(ctx, "cannot create config", zap.Error(err))
	}

	database, err := postgres.New(ctx, cfg)
	if err != nil {
		logger.Fatal(ctx, "cannot create db", zap.Error(err))
	}
	userClient, err := service.NewUserClient(cfg.UserClientDNS)
	if err != nil {
		logger.Fatal(ctx, "failed to create UserClient")
	}

	schedule_service := service.NewScheduleServer(database, userClient)
	if err != nil {
		logger.Fatal(ctx, "cannot create schedule_service", zap.Error(err))
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPCPort))
	if err != nil {
		logger.Fatal(ctx, "cannot create listener", zap.Error(err))
	}

	server := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			metadata.NewMetadataUnaryInterceptor(),
			logging.NewUnaryLoggingInterceptor(logger),
		)),
	)
	database.RegisterHealthService(server) //readiness probe
	pb.RegisterScheduleServiceServer(server, schedule_service)

	logger.Info(ctx, "Starting gRPC server...", zap.String("port", cfg.GRPCPort),
		zap.String("user_service_addr", cfg.UserClientDNS))
	go func() {
		if err := server.Serve(listener); err != nil {
			logger.Fatal(ctx, "failed to serve", zap.Error(err))
		}
	}()

	select {
	case <-ctx.Done():
		server.GracefulStop()
		database.Close()
		userClient.Close()
		logger.Info(ctx, "Server Stopped")

	}
}
