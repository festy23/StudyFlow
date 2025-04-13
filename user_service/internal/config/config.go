package config

import (
	"errors"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
)

type Config struct {
	GRPCPort            int    `env:"GRPC_PORT" env-default:"50051"`
	PostgresURL         string `env:"POSTGRES_URL" env-default:"postgres://postgres:postgres@localhost:5432/postgres"`
	PostgresMaxConn     int32  `env:"POSTGRES_MAX_CONN" env-default:"5"`
	PostgresMinConn     int32  `env:"POSTGRES_MIN_CONN" env-default:"1"`
	PostgresAutoMigrate bool   `env:"POSTGRES_AUTO_MIGRATE" env-default:"true"`
	TelegramSecret      string `env:"TELEGRAM_SECRET" env-default:"no-secret"`
}

func New() (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadConfig("./config/.env", &cfg); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if err := cleanenv.ReadEnv(&cfg); err != nil {
				return nil, err
			}
			return &cfg, nil
		}
		return nil, err
	}
	return &cfg, nil
}
