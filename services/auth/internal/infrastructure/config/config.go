package config

import (
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v3"
)

type Config struct {
	ServiceName string         `yaml:"service_name"`
	Server      ServerConfig   `yaml:"server"`
	Postgres    PostgresConfig `yaml:"postgres"`
	Auth        AuthConfig     `yaml:"auth"`
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

type AuthConfig struct {
	SessionTTL string `yaml:"session_ttl"`
	BcryptCost int    `yaml:"bcrypt_cost"`
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
	cfg.Postgres.Host = getEnv("AUTH_DB_HOST", getEnv("DB_HOST", cfg.Postgres.Host))
	cfg.Postgres.Port = getEnv("AUTH_DB_PORT", getEnv("DB_PORT", cfg.Postgres.Port))
	cfg.Postgres.User = getEnv("AUTH_DB_USER", getEnv("DB_USER", cfg.Postgres.User))
	cfg.Postgres.Name = getEnv("AUTH_DB_NAME", cfg.Postgres.Name)
	cfg.Postgres.Password = getEnv("AUTH_DB_PASSWORD", getEnv("DB_PASSWORD", "postgres"))

	return cfg, nil
}

func setDefaults(cfg *Config) {
	if cfg.ServiceName == "" {
		cfg.ServiceName = "auth-service"
	}
	if cfg.Server.Host == "" {
		cfg.Server.Host = "0.0.0.0"
	}
	if cfg.Server.GRPCPort == "" {
		cfg.Server.GRPCPort = "9091"
	}
	if cfg.Server.HTTPPort == "" {
		cfg.Server.HTTPPort = "8081"
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
		cfg.Postgres.Name = "sporttech_auth"
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
	if cfg.Auth.SessionTTL == "" {
		cfg.Auth.SessionTTL = "720h"
	}
	if cfg.Auth.BcryptCost == 0 {
		cfg.Auth.BcryptCost = bcrypt.DefaultCost
	}
	if cfg.OpenAPI.FilePath == "" {
		cfg.OpenAPI.FilePath = "grpc/gen/openapiv2/auth/v1/auth.swagger.json"
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

func (cfg AuthConfig) SessionTTLDuration() (time.Duration, error) {
	return time.ParseDuration(cfg.SessionTTL)
}
