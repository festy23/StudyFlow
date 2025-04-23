package main

import (
	"apigateway/internal/cache"
	"apigateway/internal/client"
	"apigateway/internal/config"
	"apigateway/internal/handler"
	"apigateway/internal/middleware"
	"common_library/logging"
	"context"
	filepb "fileservice/pkg/api"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	homeworkpb "homework_service/pkg/api"
	"net/http"
	paymentpb "paymentservice/pkg/api"
	schedulepb "schedule_service/pkg/api"
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

	userGrpcClient, closeFunc := client.New(ctx, cfg.UserServiceURL)
	defer closeFunc()

	fileGrpcClient, closeFunc := client.New(ctx, cfg.FileServiceURL)
	defer closeFunc()

	homeworkGrpcClient, closeFunc := client.New(ctx, cfg.HomeworkServiceURL)
	defer closeFunc()

	paymentGrpcClient, closeFunc := client.New(ctx, cfg.PaymentServiceURL)
	defer closeFunc()

	scheduleGrpcClient, closeFunc := client.New(ctx, cfg.ScheduleServiceURL)
	defer closeFunc()

	redisConn := redis.NewClient(&redis.Options{
		Addr: cfg.RedisURL,
	})

	redisCache := cache.NewRedisCache(redisConn)

	userClient := userpb.NewUserServiceClient(userGrpcClient)
	userHandler := handler.NewUserHandler(userClient, redisCache)
	authHandler := handler.NewSignUpHandler(userClient)

	fileClient := filepb.NewFileServiceClient(fileGrpcClient)
	fileHandler := handler.NewFileHandler(fileClient, cfg.MinioURL)

	homeworkClient := homeworkpb.NewHomeworkServiceClient(homeworkGrpcClient)
	homeworkHandler := handler.NewHomeworkHandler(homeworkClient)

	paymentClient := paymentpb.NewPaymentServiceClient(paymentGrpcClient)
	paymentHandler := handler.NewPaymentHandler(paymentClient)

	scheduleClient := schedulepb.NewScheduleServiceClient(scheduleGrpcClient)
	scheduleHandler := handler.NewScheduleHandler(scheduleClient)

	authMiddleware := middleware.NewAuthMiddleware(userClient)
	r := chi.NewRouter()
	r.Use(middleware.NewLoggingMiddleware(logger))
	r.Route("/users", func(r chi.Router) {
		authHandler.RegisterRoutes(r)
		userHandler.RegisterRoutes(r, authMiddleware)
	})

	r.Route("/files", func(r chi.Router) {
		fileHandler.RegisterRoutes(r)
	})

	r.Route("/schedule", func(r chi.Router) {
		scheduleHandler.RegisterRoutes(r, authMiddleware)
	})

	r.Route("/payment", func(r chi.Router) {
		paymentHandler.RegisterRoutes(r, authMiddleware)
	})

	r.Route("/homework", func(r chi.Router) {
		homeworkHandler.RegisterRoutes(r, authMiddleware)
	})

	port := fmt.Sprintf(":%d", cfg.HTTPPort)
	logger.Info(ctx, "Starting server", zap.String("port", port))

	err = http.ListenAndServe(port, r)
	if err != nil {
		logger.Fatal(ctx, "cannot start http server", zap.Error(err))
	}
}
