package main

import (
	"time"

	"payment_service/internal/app"
	"payment_service/internal/config"
)

func main() {
	cfg := config.Config{
		GRPC: config.GRPCConfig{
			Port:    50051,
			Timeout: 10 * time.Second,
		},
		Postgres: config.PostgresConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "postgres",
			DBName:   "payment_db",
			SSLMode:  "disable",
		},
		Kafka: config.KafkaConfig{
			Brokers:      []string{"localhost:9092"},
			PaymentTopic: "payment_events",
		},
		Workers: config.WorkersConfig{
			ReminderInterval: 24 * time.Hour,
		},
	}

	app := app.New(&cfg)
	app.Run()
}
