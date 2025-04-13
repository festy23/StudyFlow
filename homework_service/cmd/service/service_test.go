package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"homework_service/internal/server/grpc"
	"homework_service/pkg/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserClient struct {
	mock.Mock
}

func (m *MockUserClient) UserExists(ctx context.Context, userID string) bool {
	args := m.Called(ctx, userID)
	return args.Bool(0)
}

func (m *MockUserClient) IsPair(ctx context.Context, tutorID, studentID string) bool {
	args := m.Called(ctx, tutorID, studentID)
	return args.Bool(0)
}

func (m *MockUserClient) GetUserRole(ctx context.Context, userID string) (string, error) {
	args := m.Called(ctx, userID)
	return args.String(0), args.Error(1)
}

func TestMainFunction(t *testing.T) {
	cfg := &Config{}
	cfg.GRPC.Address = ":50052"
	cfg.Services.User = "http://mock-user-service"
	cfg.Services.File = "http://mock-file-service"
	grpcServer := grpc.NewServer(grpc.Config{Address: cfg.GRPC.Address}, nil)
	go func() {
		log := logger.New()
		grpcServer := grpc.NewServer(grpc.Config{Address: cfg.GRPC.Address}, nil)
		listener, err := net.Listen("tcp", cfg.GRPC.Address)
		if err != nil {
			log.Fatalf("Failed to listen: %v", err)
		}

		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	time.Sleep(1 * time.Second)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	quit <- syscall.SIGINT

	assert.NotNil(t, grpcServer)
}
