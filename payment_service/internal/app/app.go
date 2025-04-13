package app

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"payment_service/internal/config"
	"payment_service/internal/repository/postgres"
	"payment_service/internal/service"
	"payment_service/internal/transport/grpc"
	"payment_service/internal/worker"
)

type App struct {
	cfg        *config.Config
	grpcServer *grpc.Server
	worker     *worker.ReminderWorker
}

func New(cfg *config.Config) *App {
	db, err := postgres.NewPostgresDB(cfg.Postgres)
	if err != nil {
		log.Fatalf("failed to initialize db: %v", err)
	}
	repo := postgres.NewPaymentRepository(db)

	svc := service.NewPaymentService(repo)

	grpcServer := grpc.NewServer(svc, cfg.GRPC)

	reminderWorker := worker.NewReminderWorker(svc, cfg.Workers.ReminderInterval)

	return &App{
		cfg:        cfg,
		grpcServer: grpcServer,
		worker:     reminderWorker,
	}
}

func (a *App) Run() {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		if err := a.grpcServer.Run(); err != nil {
			log.Printf("grpc server run error: %v", err)
			cancel()
		}
	}()

	go a.worker.Start(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	select {
	case <-quit:
	case <-ctx.Done():
	}

	a.grpcServer.Stop()
	cancel()
}
