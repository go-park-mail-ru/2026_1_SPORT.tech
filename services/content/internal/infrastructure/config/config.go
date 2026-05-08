package config

import (
	"fmt"
	"net"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	ServiceName string         `yaml:"service_name" env:"SERVICE_NAME" env-default:"content-service" validate:"required"`
	Server      ServerConfig   `yaml:"server"`
	Postgres    PostgresConfig `yaml:"postgres"`
	Storage     StorageConfig  `yaml:"storage"`
	OpenAPI     OpenAPIConfig  `yaml:"openapi"`
}

type ServerConfig struct {
	Host            string `yaml:"host" env:"CONTENT_SERVER_HOST" env-default:"0.0.0.0" validate:"required"`
	GRPCPort        string `yaml:"grpc_port" env:"CONTENT_GRPC_PORT" env-default:"9093" validate:"required"`
	HTTPPort        string `yaml:"http_port" env:"CONTENT_HTTP_PORT" env-default:"8083" validate:"required"`
	ShutdownTimeout string `yaml:"shutdown_timeout" env:"CONTENT_SHUTDOWN_TIMEOUT" env-default:"10s" validate:"required"`
}

type PostgresConfig struct {
	Host            string `yaml:"host" env:"CONTENT_DB_HOST" env-default:"localhost" validate:"required"`
	Port            string `yaml:"port" env:"CONTENT_DB_PORT" env-default:"5432" validate:"required"`
	User            string `yaml:"user" env:"DB_USER" validate:"required"`
	Password        string `yaml:"password" env:"DB_PASSWORD" validate:"required"`
	Name            string `yaml:"db_name" env:"CONTENT_DB_NAME" env-default:"sporttech_content" validate:"required"`
	MaxOpenConns    int    `yaml:"max_open_conns" env:"CONTENT_DB_MAX_OPEN_CONNS" env-default:"20" validate:"required"`
	MaxIdleConns    int    `yaml:"max_idle_conns" env:"CONTENT_DB_MAX_IDLE_CONNS" env-default:"10" validate:"required"`
	ConnMaxLifetime string `yaml:"conn_max_lifetime" env:"CONTENT_DB_CONN_MAX_LIFETIME" env-default:"30m" validate:"required"`
}

type StorageConfig struct {
	Host          string `yaml:"host" env:"CONTENT_STORAGE_HOST" env-default:"localhost" validate:"required"`
	Port          string `yaml:"port" env:"CONTENT_STORAGE_PORT" env-default:"8000" validate:"required"`
	Bucket        string `yaml:"bucket" env:"CONTENT_STORAGE_BUCKET" env-default:"post-media" validate:"required"`
	PublicBaseURL string `yaml:"public_base_url" env:"CONTENT_STORAGE_PUBLIC_BASE_URL" env-default:"http://localhost:8000/post-media" validate:"required"`
	UseSSL        bool   `yaml:"use_ssl" env:"CONTENT_STORAGE_USE_SSL" env-default:"false"`
	AccessKey     string `yaml:"access_key" env:"MINIO_ACCESS_KEY" validate:"required"`
	SecretKey     string `yaml:"secret_key" env:"MINIO_SECRET_KEY" validate:"required"`
}

type OpenAPIConfig struct {
	FilePath string `yaml:"file_path" env:"CONTENT_OPENAPI_FILE_PATH" env-default:"grpc/gen/openapiv2/content/v1/content.swagger.json" validate:"required"`
}

func NewConfig(path string) (Config, error) {
	var cfg Config
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return Config{}, fmt.Errorf("read env: %w", err)
	}
	if err := validator.New().Struct(&cfg); err != nil {
		return Config{}, fmt.Errorf("validate config: %w", err)
	}

	return cfg, nil
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
