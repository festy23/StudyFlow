package config

import (
	"errors"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
)

type Config struct {
	HTTPPort           int    `env:"HTTP_PORT" env-default:"8080"`
	UserServiceURL     string `env:"USER_CLIENT_URL"`
	FileServiceURL     string `env:"FILE_SERVICE_URL"`
	ScheduleServiceURL string `env:"SCHEDULE_SERVICE_URL"`
	HomeworkServiceURL string `env:"HOMEWORK_SERVICE_URL"`
	PaymentServiceURL  string `env:"PAYMENT_SERVICE_URL"`
	MinioURL           string `env:"MINIO_URL"`
	RedisURL           string `env:"REDIS_URL"`
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
