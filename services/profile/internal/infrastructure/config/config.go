package config

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	ServiceName string         `yaml:"service_name" env:"SERVICE_NAME" env-default:"profile-service" validate:"required"`
	Server      ServerConfig   `yaml:"server"`
	Postgres    PostgresConfig `yaml:"postgres"`
	Storage     StorageConfig  `yaml:"storage"`
	OpenAPI     OpenAPIConfig  `yaml:"openapi"`
}

type ServerConfig struct {
	Host            string `yaml:"host" env:"PROFILE_SERVER_HOST" env-default:"0.0.0.0" validate:"required"`
	GRPCPort        string `yaml:"grpc_port" env:"PROFILE_GRPC_PORT" env-default:"9092" validate:"required"`
	HTTPPort        string `yaml:"http_port" env:"PROFILE_HTTP_PORT" env-default:"8082" validate:"required"`
	ShutdownTimeout string `yaml:"shutdown_timeout" env:"PROFILE_SHUTDOWN_TIMEOUT" env-default:"10s" validate:"required"`
}

type PostgresConfig struct {
	Host                              string `yaml:"host" env:"PROFILE_DB_HOST" env-default:"localhost" validate:"required"`
	Port                              string `yaml:"port" env:"PROFILE_DB_PORT" env-default:"5432" validate:"required"`
	User                              string `yaml:"user" env:"PROFILE_DB_USER" validate:"required"`
	Password                          string `yaml:"password" env:"PROFILE_DB_PASSWORD" validate:"required"`
	Name                              string `yaml:"db_name" env:"PROFILE_DB_NAME" env-default:"sporttech_profile" validate:"required"`
	ApplicationName                   string `yaml:"application_name" env:"PROFILE_DB_APPLICATION_NAME" env-default:"profile-service" validate:"required"`
	DBMaxOpenConns                    int    `yaml:"db_max_open_conns" env:"PROFILE_DB_MAX_OPEN_CONNS" env-default:"12" validate:"required"`
	DBMaxIdleConns                    int    `yaml:"db_max_idle_conns" env:"PROFILE_DB_MAX_IDLE_CONNS" env-default:"6" validate:"gte=0"`
	DBConnMaxLifetime                 string `yaml:"db_conn_max_lifetime" env:"PROFILE_DB_CONN_MAX_LIFETIME" env-default:"30m" validate:"required"`
	DBConnectTimeoutSeconds           int    `yaml:"db_connect_timeout_seconds" env:"PROFILE_DB_CONNECT_TIMEOUT_SECONDS" env-default:"5" validate:"required"`
	DBStatementTimeout                string `yaml:"db_statement_timeout" env:"PROFILE_DB_STATEMENT_TIMEOUT" env-default:"5s" validate:"required"`
	DBLockTimeout                     string `yaml:"db_lock_timeout" env:"PROFILE_DB_LOCK_TIMEOUT" env-default:"1s" validate:"required"`
	DBIdleInTransactionSessionTimeout string `yaml:"db_idle_in_transaction_session_timeout" env:"PROFILE_DB_IDLE_IN_TRANSACTION_SESSION_TIMEOUT" env-default:"10s" validate:"required"`
}

type StorageConfig struct {
	Host          string `yaml:"host" env:"PROFILE_STORAGE_HOST" env-default:"localhost" validate:"required"`
	Port          string `yaml:"port" env:"PROFILE_STORAGE_PORT" env-default:"8000" validate:"required"`
	Bucket        string `yaml:"bucket" env:"PROFILE_STORAGE_BUCKET" env-default:"avatars" validate:"required"`
	PublicBaseURL string `yaml:"public_base_url" env:"PROFILE_STORAGE_PUBLIC_BASE_URL" env-default:"http://localhost:8000/avatars" validate:"required"`
	UseSSL        bool   `yaml:"use_ssl" env:"PROFILE_STORAGE_USE_SSL" env-default:"false"`
	AccessKey     string `yaml:"access_key" env:"MINIO_ACCESS_KEY" validate:"required"`
	SecretKey     string `yaml:"secret_key" env:"MINIO_SECRET_KEY" validate:"required"`
}

type OpenAPIConfig struct {
	FilePath string `yaml:"file_path" env:"PROFILE_OPENAPI_FILE_PATH" env-default:"grpc/gen/openapiv2/profile/v1/profile.swagger.json" validate:"required"`
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
	if err := cfg.Postgres.Validate(); err != nil {
		return Config{}, fmt.Errorf("validate postgres config: %w", err)
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
	query := url.Values{}
	query.Set("sslmode", "disable")
	query.Set("application_name", cfg.ApplicationName)
	query.Set("connect_timeout", strconv.Itoa(cfg.DBConnectTimeoutSeconds))
	query.Set("statement_timeout", cfg.DBStatementTimeout)
	query.Set("lock_timeout", cfg.DBLockTimeout)
	query.Set("idle_in_transaction_session_timeout", cfg.DBIdleInTransactionSessionTimeout)

	databaseURL := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     net.JoinHostPort(cfg.Host, cfg.Port),
		Path:     cfg.Name,
		RawQuery: query.Encode(),
	}

	return databaseURL.String()
}

func (cfg PostgresConfig) DBConnMaxLifetimeDuration() (time.Duration, error) {
	return time.ParseDuration(cfg.DBConnMaxLifetime)
}

func (cfg PostgresConfig) Validate() error {
	if cfg.DBMaxOpenConns <= 0 {
		return fmt.Errorf("max_open_conns must be positive")
	}
	if cfg.DBMaxIdleConns > cfg.DBMaxOpenConns {
		return fmt.Errorf("max_idle_conns must not exceed max_open_conns")
	}
	if cfg.DBConnectTimeoutSeconds <= 0 {
		return fmt.Errorf("connect_timeout_seconds must be positive")
	}

	statementTimeout, err := parsePositiveDuration("statement_timeout", cfg.DBStatementTimeout)
	if err != nil {
		return err
	}
	lockTimeout, err := parsePositiveDuration("lock_timeout", cfg.DBLockTimeout)
	if err != nil {
		return err
	}
	if lockTimeout >= statementTimeout {
		return fmt.Errorf("lock_timeout must be less than statement_timeout")
	}
	if _, err := parsePositiveDuration("idle_in_transaction_session_timeout", cfg.DBIdleInTransactionSessionTimeout); err != nil {
		return err
	}
	if _, err := parsePositiveDuration("conn_max_lifetime", cfg.DBConnMaxLifetime); err != nil {
		return err
	}

	return nil
}

func parsePositiveDuration(name string, value string) (time.Duration, error) {
	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", name, err)
	}
	if duration <= 0 {
		return 0, fmt.Errorf("%s must be positive", name)
	}

	return duration, nil
}

func (cfg StorageConfig) Endpoint() string {
	return net.JoinHostPort(cfg.Host, cfg.Port)
}
