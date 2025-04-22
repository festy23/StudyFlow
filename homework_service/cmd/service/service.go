package main

import (
	"common_library/logging"
	"common_library/metadata"
	"google.golang.org/grpc/credentials/insecure"
	configs "homework_service/config"
	"net"
	"os"
	"os/signal"
	"syscall"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"

	"homework_service/internal/app"
	"homework_service/internal/repository"
	"homework_service/internal/server/homework_grpc"
	"homework_service/internal/service"
	"homework_service/pkg/db"
	"homework_service/pkg/kafka"
	"homework_service/pkg/logger"

	_ "github.com/lib/pq"
)

func main() {
	log := logger.New()

	cfg, err := configs.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	dbConfig := db.Config{
		Host:     cfg.DB.Host,
		Port:     cfg.DB.Port,
		User:     cfg.DB.User,
		Password: cfg.DB.Password,
		DBName:   cfg.DB.DBName,
		SSLMode:  cfg.DB.SSLMode,
	}

	pg, err := db.NewPostgres(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pg.Close()

	assignmentRepo := repository.NewAssignmentRepository(pg.DB())
	submissionRepo := repository.NewSubmissionRepository(pg.DB())
	feedbackRepo := repository.NewFeedbackRepository(pg.DB())

	userGrpc, err := grpc.NewClient(
		cfg.Services.UserService.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Failed to create user service: %v", err)
	}
	fileGrpc, err := grpc.NewClient(
		cfg.Services.FileService.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Failed to create file service: %v", err)
	}
	userClient := app.NewUserClient(userGrpc)
	fileClient := app.NewFileClient(fileGrpc)

	assignmentService := service.NewAssignmentService(
		*assignmentRepo,
		userClient,
		fileClient,
	)

	submissionService := service.NewSubmissionService(
		submissionRepo,
		assignmentRepo,
		fileClient,
	)

	feedbackService := service.NewFeedbackService(
		feedbackRepo,
		submissionRepo,
		assignmentRepo,
		fileClient,
	)

	handler := homework_grpc.NewHomeworkHandler(
		*assignmentService,
		submissionService,
		feedbackService,
		log,
	)

	kafkaConfig := kafka.Config{
		Brokers: cfg.Kafka.Brokers,
	}

	kafkaProducer, err := kafka.NewProducer(kafkaConfig)
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer kafkaProducer.Close()

	interceptor := grpc_middleware.ChainUnaryServer(
		metadata.NewMetadataUnaryInterceptor(),
		logging.NewUnaryLoggingInterceptor(logging.New(log.ZapLogger)),
	)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(interceptor),
	)

	homework_grpc.RegisterHomeworkServiceServer(grpcServer, handler)

	listener, err := net.Listen("tcp", cfg.GRPC.Address)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	go func() {
		log.Infof("Starting gRPC server on %s", cfg.GRPC.Address)
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")
	grpcServer.GracefulStop()
	log.Info("Server stopped")
}
