package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"
	"homework_service/internal/app"
	"homework_service/internal/repository"
	"homework_service/internal/server/grpc"
	"homework_service/internal/service"
	"homework_service/pkg/db"
	"homework_service/pkg/kafka"
	"homework_service/pkg/logger"
	"homework_service/proto/homework/v1"
)

type UserClient interface {
	UserExists(ctx context.Context, userID string) bool
	IsPair(ctx context.Context, tutorID, studentID string) bool
	GetUserRole(ctx context.Context, userID string) (string, error)
}

func main() {
	log := logger.New()

	cfg, err := LoadConfig()
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

	userClient := app.NewUserClient(cfg.Services.User)
	fileClient := app.NewFileClient(cfg.Services.File)

	assignmentService := service.NewAssignmentService(assignmentRepo, userClient, fileClient)
	submissionService := service.NewSubmissionService(submissionRepo, assignmentRepo, fileClient)
	feedbackService := service.NewFeedbackService(feedbackRepo, submissionRepo, assignmentRepo, fileClient)

	kafkaConfig := kafka.Config{
		Brokers: cfg.Kafka.Brokers,
	}

	kafkaProducer, err := kafka.NewProducer(kafkaConfig)
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer kafkaProducer.Close()

	handler := grpc.NewHomeworkHandler(
		assignmentService,
		submissionService,
		feedbackService,
	)
	grpcServer := grpc.NewServer(grpc.Config{Address: cfg.GRPC.Address}, handler)

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
	grpcServer.Stop()
	log.Info("Server stopped")
}

type Config struct {
	DB struct {
		Host     string
		Port     int
		User     string
		Password string
		DBName   string
		SSLMode  string
	}
	Kafka struct {
		Brokers []string
	}
	GRPC struct {
		Address string
	}
	Services struct {
		User string
		File string
	}
}

func LoadConfig() (*Config, error) {
	return &Config{
		DB: struct {
			Host     string
			Port     int
			User     string
			Password string
			DBName   string
			SSLMode  string
		}{
			Host:     "localhost",
			Port:     5432,
			User:     "user",
			Password: "password",
			DBName:   "homework",
			SSLMode:  "disable",
		},
		Kafka: struct {
			Brokers []string
		}{
			Brokers: []string{"localhost:9092"},
		},
		GRPC: struct {
			Address string
		}{
			Address: ":50051",
		},
		Services: struct {
			User string
			File string
		}{
			User: "http://user-service",
			File: "http://file-service",
		},
	}, nil
}
