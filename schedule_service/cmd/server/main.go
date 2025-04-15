// package main

// import (
// 	"common_library/interceptors"
// 	"context"
// 	"fmt"
// 	"net"
// 	"os"
// 	"os/signal"
// 	"schedule_service/internal/config"
// 	"schedule_service/internal/data"
// 	"schedule_service/internal/db"
// 	service "schedule_service/internal/services/lesson"
// 	pb "schedule_service/pkg/api"
// 	"syscall"

// 	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
// 	"go.uber.org/zap"
// 	"google.golang.org/grpc"
// )

// func main() {
// 	ctx := context.Background()
// 	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
// 	defer stop()

// 	zapLogger, err := zap.NewDevelopment()
// 	if err != nil {
// 		panic(err)
// 	}

// 	logger := interceptors.lNew(zapLogger)

// 	ctx = logging.ContextWithLogger(ctx, logger)

// 	cfg, err := config.New()
// 	if err != nil {
// 		logger.Fatal(ctx, "cannot create config", zap.Error(err))
// 	}

// 	database, err := db.New(ctx, cfg)
// 	if err != nil {
// 		logger.Fatal(ctx, "cannot create db", zap.Error(err))
// 	}

// 	repo := data.NewFileRepository(database)

// 	s3Client, err := s3_client.New(ctx, cfg)
// 	if err != nil {
// 		logger.Fatal(ctx, "cannot create S3 client", zap.Error(err))
// 	}

// 	schedule_service, err := service.Newschedule_service(ctx, repo, s3Client, "user-files")
// 	if err != nil {
// 		logger.Fatal(ctx, "cannot create schedule_service", zap.Error(err))
// 	}

// 	fileHandler := handler.NewFileHandler(schedule_service)

// 	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
// 	if err != nil {
// 		logger.Fatal(ctx, "cannot create listener", zap.Error(err))
// 	}

// 	server := grpc.NewServer(
// 		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
// 			metadata.NewMetadataUnaryInterceptor(),
// 			logging.NewUnaryLoggingInterceptor(logger),
// 		)),
// 	)

// 	pb.Registerschedule_serviceServer(server, fileHandler)

// 	logger.Info(ctx, "Starting gRPC server...", zap.Int("port", cfg.GRPCPort))
// 	go func() {
// 		if err := server.Serve(listener); err != nil {
// 			logger.Fatal(ctx, "failed to serve", zap.Error(err))
// 		}
// 	}()

// 	select {
// 	case <-ctx.Done():
// 		server.Stop()
// 		logger.Info(ctx, "Server Stopped")
// 	}
// }
