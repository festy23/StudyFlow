package config

import (
	"log"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	GRPCPort            string `env:"GRPC_PORT" env-required:"true"`
	PostgresURL         string `env:"POSTGRES_URL" env-required:"true"`
	PostgresMaxConn     int    `env:"POSTGRES_MAX_CONN" env-default:"5"`
	PostgresMinConn     int    `env:"POSTGRES_MIN_CONN" env-default:"1"`
	PostgresAutoMigrate bool   `env:"POSTGRES_AUTO_MIGRATE" env-default:"false"`
}

var (
	cfg  *Config
	once sync.Once
)

func GetConfig() *Config {
	once.Do(func() {
		cfg = &Config{}
		if err := cleanenv.ReadEnv(cfg); err != nil {
			log.Fatalf("failed to read config: %v", err)
		}
	})
	return cfg
}
