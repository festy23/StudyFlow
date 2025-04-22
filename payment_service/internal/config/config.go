package config

import (
	"time"
)

type Config struct {
	GRPC     GRPCConfig
	Postgres PostgresConfig
	Kafka    KafkaConfig
	Workers  WorkersConfig
}

type GRPCConfig struct {
	Port    int
	Timeout time.Duration
}

type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type KafkaConfig struct {
	Brokers      []string
	PaymentTopic string
}

type WorkersConfig struct {
	ReminderInterval time.Duration
}
