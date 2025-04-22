package config

import (
	"errors"
	"log"
	"os"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	GRPCPort            string `env:"GRPC_PORT" env-required:"true"`
	PostgresURL         string `env:"POSTGRES_URL" env-default:"postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"`
	PostgresMaxConn     int    `env:"POSTGRES_MAX_CONN" env-default:"5"`
	PostgresMinConn     int    `env:"POSTGRES_MIN_CONN" env-default:"1"`
	PostgresAutoMigrate bool   `env:"POSTGRES_AUTO_MIGRATE" env-default:"false"`
	UserClientDNS       string `env:"USER_CLIENT_DNS" env-required:"true"`
}

var (
	cfg  *Config
	once sync.Once
)

func GetConfig() *Config {
	once.Do(func() {
		cfg = &Config{}
		// .env должен быть в корне (на одном уровне с go.mod)
		if err := cleanenv.ReadConfig(".env", cfg); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				if err := cleanenv.ReadEnv(cfg); err != nil {
					log.Fatalf("failed to read config: %v", err)
				}
				return
			}
			help, _ := cleanenv.GetDescription(cfg, nil)
			log.Printf("Config help:\n%s", help)
			log.Fatalf("failed to read config: %v", err)
		}
	})
	return cfg
}
