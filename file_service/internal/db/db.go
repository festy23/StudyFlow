package db

import (
	"context"
	"errors"
	"fileservice/internal/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

func New(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	if cfg.PostgresAutoMigrate {
		if err := runMigrations(cfg); err != nil {
			return nil, err
		}
	}

	pgxCfg, err := pgxpool.ParseConfig(cfg.PostgresURL)
	if err != nil {
		return nil, err
	}

	pgxCfg.MaxConns = cfg.PostgresMaxConn
	pgxCfg.MinConns = cfg.PostgresMinConn

	pool, err := pgxpool.NewWithConfig(ctx, pgxCfg)
	if err != nil {
		return nil, err
	}

	err = pool.Ping(ctx)
	if err != nil {
		return nil, err
	}

	return pool, nil
}

func runMigrations(cfg *config.Config) error {
	m, err := migrate.New(
		"file://migrations",
		cfg.PostgresURL,
	)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}
