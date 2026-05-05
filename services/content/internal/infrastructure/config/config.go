package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ServiceName string         `yaml:"service_name"`
	Server      ServerConfig   `yaml:"server"`
	Postgres    PostgresConfig `yaml:"postgres"`
	Storage     StorageConfig  `yaml:"storage"`
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

type StorageConfig struct {
	Host          string `yaml:"host"`
	Port          string `yaml:"port"`
	Bucket        string `yaml:"bucket"`
	PublicBaseURL string `yaml:"public_base_url"`
	UseSSL        bool   `yaml:"use_ssl"`
	AccessKey     string
	SecretKey     string
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
	cfg.Postgres.Host = getEnv("CONTENT_DB_HOST", getEnv("DB_HOST", cfg.Postgres.Host))
	cfg.Postgres.Port = getEnv("CONTENT_DB_PORT", getEnv("DB_PORT", cfg.Postgres.Port))
	cfg.Postgres.User = getEnv("CONTENT_DB_USER", getEnv("DB_USER", cfg.Postgres.User))
	cfg.Postgres.Name = getEnv("CONTENT_DB_NAME", cfg.Postgres.Name)
	cfg.Postgres.Password = getEnv("CONTENT_DB_PASSWORD", getEnv("DB_PASSWORD", "postgres"))
	cfg.Storage.Host = getEnv("CONTENT_STORAGE_HOST", getEnv("STORAGE_HOST", cfg.Storage.Host))
	cfg.Storage.Port = getEnv("CONTENT_STORAGE_PORT", getEnv("STORAGE_PORT", cfg.Storage.Port))
	cfg.Storage.Bucket = getEnv("CONTENT_STORAGE_BUCKET", getEnv("STORAGE_BUCKET", cfg.Storage.Bucket))
	cfg.Storage.PublicBaseURL = getEnv("CONTENT_STORAGE_PUBLIC_BASE_URL", getEnv("STORAGE_PUBLIC_BASE_URL", cfg.Storage.PublicBaseURL))
	cfg.Storage.UseSSL = getEnvBool("CONTENT_STORAGE_USE_SSL", getEnvBool("STORAGE_USE_SSL", cfg.Storage.UseSSL))
	cfg.Storage.AccessKey = getEnv("MINIO_ACCESS_KEY", "minioadmin")
	cfg.Storage.SecretKey = getEnv("MINIO_SECRET_KEY", "minioadmin")

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
	if cfg.Storage.Host == "" {
		cfg.Storage.Host = "localhost"
	}
	if cfg.Storage.Port == "" {
		cfg.Storage.Port = "8000"
	}
	if cfg.Storage.Bucket == "" {
		cfg.Storage.Bucket = "post-media"
	}
	if cfg.Storage.PublicBaseURL == "" {
		cfg.Storage.PublicBaseURL = "http://localhost:8000/post-media"
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

func getEnvBool(key string, fallback bool) bool {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	parsedValue, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return parsedValue
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

func (cfg StorageConfig) Endpoint() string {
	return net.JoinHostPort(cfg.Host, cfg.Port)
}
