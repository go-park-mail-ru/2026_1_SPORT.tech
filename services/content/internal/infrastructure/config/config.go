package config

import (
	"fmt"
	"net"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ServiceName string         `yaml:"service_name"`
	Server      ServerConfig   `yaml:"server"`
	Postgres    PostgresConfig `yaml:"postgres"`
	OpenAPI     OpenAPIConfig  `yaml:"openapi"`
}

type ServerConfig struct {
	Host            string `yaml:"host"`
	GRPCPort        string `yaml:"grpc_port"`
	HTTPPort        string `yaml:"http_port"`
	ShutdownTimeout string `yaml:"shutdown_timeout"`
}

type PostgresConfig struct {
	Host            string `yaml:"host"`
	Port            string `yaml:"port"`
	User            string `yaml:"user"`
	Password        string
	Name            string `yaml:"db_name"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	ConnMaxLifetime string `yaml:"conn_max_lifetime"`
}

type OpenAPIConfig struct {
	FilePath string `yaml:"file_path"`
}

func NewConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}

	setDefaults(&cfg)
	cfg.Postgres.Password = getEnv("CONTENT_DB_PASSWORD", getEnv("DB_PASSWORD", "postgres"))

	return cfg, nil
}

func setDefaults(cfg *Config) {
	if cfg.ServiceName == "" {
		cfg.ServiceName = "content-service"
	}
	if cfg.Server.Host == "" {
		cfg.Server.Host = "0.0.0.0"
	}
	if cfg.Server.GRPCPort == "" {
		cfg.Server.GRPCPort = "9093"
	}
	if cfg.Server.HTTPPort == "" {
		cfg.Server.HTTPPort = "8083"
	}
	if cfg.Server.ShutdownTimeout == "" {
		cfg.Server.ShutdownTimeout = "10s"
	}
	if cfg.Postgres.Host == "" {
		cfg.Postgres.Host = "localhost"
	}
	if cfg.Postgres.Port == "" {
		cfg.Postgres.Port = "5432"
	}
	if cfg.Postgres.User == "" {
		cfg.Postgres.User = "postgres"
	}
	if cfg.Postgres.Name == "" {
		cfg.Postgres.Name = "sporttech_content"
	}
	if cfg.Postgres.MaxOpenConns == 0 {
		cfg.Postgres.MaxOpenConns = 20
	}
	if cfg.Postgres.MaxIdleConns == 0 {
		cfg.Postgres.MaxIdleConns = 10
	}
	if cfg.Postgres.ConnMaxLifetime == "" {
		cfg.Postgres.ConnMaxLifetime = "30m"
	}
	if cfg.OpenAPI.FilePath == "" {
		cfg.OpenAPI.FilePath = "grpc/gen/openapiv2/content/v1/content.swagger.json"
	}
}

func getEnv(key string, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	return value
}

func (cfg ServerConfig) GRPCAddress() string {
	return net.JoinHostPort(cfg.Host, cfg.GRPCPort)
}

func (cfg ServerConfig) HTTPAddress() string {
	return net.JoinHostPort(cfg.Host, cfg.HTTPPort)
}

func (cfg ServerConfig) ShutdownTimeoutDuration() (time.Duration, error) {
	return time.ParseDuration(cfg.ShutdownTimeout)
}

func (cfg PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Name,
	)
}

func (cfg PostgresConfig) ConnMaxLifetimeDuration() (time.Duration, error) {
	return time.ParseDuration(cfg.ConnMaxLifetime)
}
