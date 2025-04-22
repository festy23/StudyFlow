package configs

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	GRPC     GRPCConfig  `yaml:"grpc"`
	DB       DBConfig    `yaml:"db"`
	Kafka    KafkaConfig `yaml:"kafka"`
	Services Services    `yaml:"services"`
}

type GRPCConfig struct {
	Address string        `yaml:"address"`
	Timeout time.Duration `yaml:"timeout"`
}

type DBConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

type KafkaConfig struct {
	Brokers        []string `yaml:"brokers"`
	Topic          string   `yaml:"topic"`
	GroupID        string   `yaml:"group_id"`
	WorkerPoolSize int      `yaml:"worker_pool_size"`
}

type Services struct {
	UserService ServiceConfig `yaml:"user_service"`
	FileService ServiceConfig `yaml:"file_service"`
}

type ServiceConfig struct {
	Address string        `yaml:"address"`
	Timeout time.Duration `yaml:"timeout"`
}

func Load() (*Config, error) {
	configPath := getConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	setDefaults(&cfg)
	overrideFromEnv(&cfg)

	if err := validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

func getConfigPath() string {
	if path := os.Getenv("CONFIG_PATH"); path != "" {
		return path
	}

	possiblePaths := []string{
		"config/config.yaml",
		"/etc/homework-service/config.yaml",
		"./config.yaml",
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return "config.yaml"
}

func setDefaults(cfg *Config) {
	if cfg.GRPC.Timeout == 0 {
		cfg.GRPC.Timeout = 30 * time.Second
	}

	if cfg.Services.UserService.Timeout == 0 {
		cfg.Services.UserService.Timeout = 10 * time.Second
	}

	if cfg.Services.FileService.Timeout == 0 {
		cfg.Services.FileService.Timeout = 10 * time.Second
	}

	if cfg.Kafka.WorkerPoolSize == 0 {
		cfg.Kafka.WorkerPoolSize = 5
	}

	if cfg.Kafka.GroupID == "" {
		cfg.Kafka.GroupID = "homework-service-group"
	}
}

func overrideFromEnv(cfg *Config) {
	if val := os.Getenv("GRPC_ADDRESS"); val != "" {
		cfg.GRPC.Address = val
	}
	if val := os.Getenv("GRPC_TIMEOUT"); val != "" {
		if timeout, err := strconv.Atoi(val); err == nil {
			cfg.GRPC.Timeout = time.Duration(timeout) * time.Second
		}
	}

	if val := os.Getenv("DB_HOST"); val != "" {
		cfg.DB.Host = val
	}
	if val := os.Getenv("DB_PORT"); val != "" {
		if port, err := strconv.Atoi(val); err == nil {
			cfg.DB.Port = port
		}
	}
	if val := os.Getenv("DB_USER"); val != "" {
		cfg.DB.User = val
	}
	if val := os.Getenv("DB_PASSWORD"); val != "" {
		cfg.DB.Password = val
	}
	if val := os.Getenv("DB_NAME"); val != "" {
		cfg.DB.DBName = val
	}
	if val := os.Getenv("DB_SSL_MODE"); val != "" {
		cfg.DB.SSLMode = val
	}

	if val := os.Getenv("KAFKA_BROKERS"); val != "" {
		cfg.Kafka.Brokers = strings.Split(val, ",")
	}
	if val := os.Getenv("KAFKA_TOPIC"); val != "" {
		cfg.Kafka.Topic = val
	}
	if val := os.Getenv("KAFKA_GROUP_ID"); val != "" {
		cfg.Kafka.GroupID = val
	}
	if val := os.Getenv("KAFKA_WORKER_POOL_SIZE"); val != "" {
		if size, err := strconv.Atoi(val); err == nil {
			cfg.Kafka.WorkerPoolSize = size
		}
	}

	if val := os.Getenv("USER_SERVICE_ADDRESS"); val != "" {
		cfg.Services.UserService.Address = val
	}
	if val := os.Getenv("USER_SERVICE_TIMEOUT"); val != "" {
		if timeout, err := strconv.Atoi(val); err == nil {
			cfg.Services.UserService.Timeout = time.Duration(timeout) * time.Second
		}
	}
	if val := os.Getenv("FILE_SERVICE_ADDRESS"); val != "" {
		cfg.Services.FileService.Address = val
	}
	if val := os.Getenv("FILE_SERVICE_TIMEOUT"); val != "" {
		if timeout, err := strconv.Atoi(val); err == nil {
			cfg.Services.FileService.Timeout = time.Duration(timeout) * time.Second
		}
	}
}

func validateConfig(cfg *Config) error {
	if cfg.GRPC.Address == "" {
		return fmt.Errorf("GRPC address must be set")
	}

	if len(cfg.Kafka.Brokers) == 0 {
		return fmt.Errorf("at least one Kafka broker must be specified")
	}

	if cfg.DB.Host == "" || cfg.DB.User == "" || cfg.DB.DBName == "" {
		return fmt.Errorf("database configuration is incomplete")
	}

	if cfg.Services.UserService.Address == "" {
		return fmt.Errorf("user service address must be specified")
	}

	if cfg.Services.FileService.Address == "" {
		return fmt.Errorf("file service address must be specified")
	}

	return nil
}

func (c *Config) GetDBConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.DB.Host,
		c.DB.Port,
		c.DB.User,
		c.DB.Password,
		c.DB.DBName,
		c.DB.SSLMode,
	)
}
