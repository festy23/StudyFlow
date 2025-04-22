package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"time"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type Postgres struct {
	db *sql.DB
}

func NewPostgres(cfg Config) (*Postgres, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open db connection: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		cfg.DBName,
		driver,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to init migration: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	return &Postgres{db: db}, nil
}

func (p *Postgres) Close() error {
	return p.db.Close()
}

func (p *Postgres) DB() *sql.DB {
	return p.db
}
