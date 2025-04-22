package main

import (
	"apigateway/internal/config"
	"apigateway/internal/handler"
	"apigateway/internal/middleware"
	"common_library/logging"
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net/http"
	userpb "userservice/pkg/api"
)

func main() {
	ctx := context.Background()
	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	logger := logging.New(zapLogger)

	cfg, err := config.New()
	if err != nil {
		logger.Fatal(ctx, "cannot create config", zap.Error(err))
	}

	userGrpcClient, err := grpc.NewClient(
		cfg.UserServiceURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logger.Fatal(ctx, "cannot create user grpc client", zap.Error(err))
	}

	defer func(userGrpcClient *grpc.ClientConn) {
		err := userGrpcClient.Close()
		if err != nil {
			logger.Fatal(ctx, "cannot close user grpc client", zap.Error(err))
		}
	}(userGrpcClient)

	userClient := userpb.NewUserServiceClient(userGrpcClient)

	userHandler := handler.NewUserHandler(userClient)
	authHandler := handler.NewSignUpHandler(userClient)

	authMiddleware := middleware.NewAuthMiddleware(userClient)
	r := chi.NewRouter()
	r.Use(middleware.NewLoggingMiddleware(logger))
	r.Route("/users", func(r chi.Router) {
		authHandler.RegisterRoutes(r)
		userHandler.RegisterRoutes(r, authMiddleware)
	})

	port := fmt.Sprintf(":%d", cfg.HTTPPort)
	logger.Info(ctx, "Starting server", zap.String("port", port))

	err = http.ListenAndServe(port, r)
	if err != nil {
		logger.Fatal(ctx, "cannot start http server", zap.Error(err))
	}
}
