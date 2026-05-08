package config

import (
	"fmt"
	"net"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
	"golang.org/x/crypto/bcrypt"
)

type Config struct {
	ServiceName string         `yaml:"service_name" env:"SERVICE_NAME" env-default:"auth-service" validate:"required"`
	Server      ServerConfig   `yaml:"server"`
	Postgres    PostgresConfig `yaml:"postgres"`
	Auth        AuthConfig     `yaml:"auth"`
	OpenAPI     OpenAPIConfig  `yaml:"openapi"`
}

type ServerConfig struct {
	Host            string `yaml:"host" env:"AUTH_SERVER_HOST" env-default:"0.0.0.0" validate:"required"`
	GRPCPort        string `yaml:"grpc_port" env:"AUTH_GRPC_PORT" env-default:"9091" validate:"required"`
	HTTPPort        string `yaml:"http_port" env:"AUTH_HTTP_PORT" env-default:"8081" validate:"required"`
	ShutdownTimeout string `yaml:"shutdown_timeout" env:"AUTH_SHUTDOWN_TIMEOUT" env-default:"10s" validate:"required"`
}

type PostgresConfig struct {
	Host            string `yaml:"host" env:"AUTH_DB_HOST" env-default:"localhost" validate:"required"`
	Port            string `yaml:"port" env:"AUTH_DB_PORT" env-default:"5432" validate:"required"`
	User            string `yaml:"user" env:"DB_USER" validate:"required"`
	Password        string `yaml:"password" env:"DB_PASSWORD" validate:"required"`
	Name            string `yaml:"db_name" env:"AUTH_DB_NAME" env-default:"sporttech_auth" validate:"required"`
	MaxOpenConns    int    `yaml:"max_open_conns" env:"AUTH_DB_MAX_OPEN_CONNS" env-default:"20" validate:"required"`
	MaxIdleConns    int    `yaml:"max_idle_conns" env:"AUTH_DB_MAX_IDLE_CONNS" env-default:"10" validate:"required"`
	ConnMaxLifetime string `yaml:"conn_max_lifetime" env:"AUTH_DB_CONN_MAX_LIFETIME" env-default:"30m" validate:"required"`
}

type AuthConfig struct {
	SessionTTL string `yaml:"session_ttl" env:"AUTH_SESSION_TTL" env-default:"720h" validate:"required"`
	BcryptCost int    `yaml:"bcrypt_cost" env:"AUTH_BCRYPT_COST" validate:"required"`
}

type OpenAPIConfig struct {
	FilePath string `yaml:"file_path" env:"AUTH_OPENAPI_FILE_PATH" env-default:"grpc/gen/openapiv2/auth/v1/auth.swagger.json" validate:"required"`
}

func NewConfig(path string) (Config, error) {
	var cfg Config
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return Config{}, fmt.Errorf("read env: %w", err)
	}
	if cfg.Auth.BcryptCost == 0 {
		cfg.Auth.BcryptCost = bcrypt.DefaultCost
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

func (cfg AuthConfig) SessionTTLDuration() (time.Duration, error) {
	return time.ParseDuration(cfg.SessionTTL)
}
