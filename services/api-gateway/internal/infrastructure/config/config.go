package config

import (
	"fmt"
	"net"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	ServiceName string           `yaml:"service_name" env:"SERVICE_NAME" env-default:"api-gateway" validate:"required"`
	Server      ServerConfig     `yaml:"server"`
	Downstream  DownstreamConfig `yaml:"downstream"`
	OpenAPI     OpenAPIConfig    `yaml:"openapi"`
}

type ServerConfig struct {
	Host            string `yaml:"host" env:"API_GATEWAY_SERVER_HOST" env-default:"0.0.0.0" validate:"required"`
	GRPCPort        string `yaml:"grpc_port" env:"API_GATEWAY_GRPC_PORT" env-default:"9090" validate:"required"`
	HTTPPort        string `yaml:"http_port" env:"API_GATEWAY_HTTP_PORT" env-default:"8080" validate:"required"`
	ShutdownTimeout string `yaml:"shutdown_timeout" env:"API_GATEWAY_SHUTDOWN_TIMEOUT" env-default:"10s" validate:"required"`
}

type DownstreamConfig struct {
	AuthGRPCEndpoint    string `yaml:"auth_grpc_endpoint" env:"API_GATEWAY_AUTH_GRPC_ENDPOINT" env-default:"localhost:9091" validate:"required"`
	ProfileGRPCEndpoint string `yaml:"profile_grpc_endpoint" env:"API_GATEWAY_PROFILE_GRPC_ENDPOINT" env-default:"localhost:9092" validate:"required"`
	ContentGRPCEndpoint string `yaml:"content_grpc_endpoint" env:"API_GATEWAY_CONTENT_GRPC_ENDPOINT" env-default:"localhost:9093" validate:"required"`
}

type OpenAPIConfig struct {
	GatewayFilePath string `yaml:"gateway_file_path" env:"API_GATEWAY_OPENAPI_FILE_PATH" env-default:"grpc/gen/openapiv2/gateway/v1/gateway.swagger.json" validate:"required"`
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

func (cfg ServerConfig) HTTPAddress() string {
	return net.JoinHostPort(cfg.Host, cfg.HTTPPort)
}

func (cfg ServerConfig) GRPCListenAddress() string {
	return net.JoinHostPort(cfg.Host, cfg.GRPCPort)
}

func (cfg ServerConfig) GRPCDialAddress() string {
	host := cfg.Host
	switch host {
	case "", "0.0.0.0", "::":
		host = "127.0.0.1"
	}

	return net.JoinHostPort(host, cfg.GRPCPort)
}

func (cfg ServerConfig) ShutdownTimeoutDuration() (time.Duration, error) {
	return time.ParseDuration(cfg.ShutdownTimeout)
}
