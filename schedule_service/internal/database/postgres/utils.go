package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgx/v5/pgxpool"

	"schedule_service/internal/config"
)

func (r *PostgresRepository) RegisterHealthService(srv *grpc.Server) {
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(srv, healthServer)

	healthServer.SetServingStatus("postgres", grpc_health_v1.HealthCheckResponse_SERVING)

	go r.watchDBConnection(healthServer)
}

func (r *PostgresRepository) watchDBConnection(healthServer *health.Server) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := r.pool.Ping(ctx); err != nil {
			healthServer.SetServingStatus("postgres", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
		} else {
			healthServer.SetServingStatus("postgres", grpc_health_v1.HealthCheckResponse_SERVING)
		}
	}
}

func New(ctx context.Context, cfg *config.Config) (*PostgresRepository, error) {
	if cfg.PostgresAutoMigrate {
		if err := runMigrations(cfg); err != nil {
			return nil, fmt.Errorf("failed to run migrations: %w", err)
		}
	}

	pool, err := createPool(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	return &PostgresRepository{pool: pool}, nil
}

func createPool(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	pgxCfg, err := pgxpool.ParseConfig(cfg.PostgresURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	pgxCfg.MaxConns = int32(cfg.PostgresMaxConn)
	pgxCfg.MinConns = int32(cfg.PostgresMinConn)

	pool, err := pgxpool.NewWithConfig(ctx, pgxCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}

func runMigrations(cfg *config.Config) error {
	m, err := migrate.New(
		"file://migrations",
		cfg.PostgresURL,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize migrations: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}

func (r *PostgresRepository) Close() {
	if r.pool != nil {
		r.pool.Close()
	}
}
