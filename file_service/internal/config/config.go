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
	S3AccessKeyID       string `env:"S3_ACCESS_KEY_ID" env-default:""`
	S3SecretAccessKey   string `env:"S3_SECRET_ACCESS_KEY" env-default:""`
	S3Endpoint          string `env:"S3_ENDPOINT" env-default:""`
	S3Region            string `env:"S3_REGION" env-default:"us-east-1"`
	GatewayPublicUrl    string `env:"GATEWAY_PUBLIC_URL" env-default:"localhost:8080"`
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
