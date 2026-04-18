package config

import (
	"fmt"
	"net"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ServiceName string           `yaml:"service_name"`
	Server      ServerConfig     `yaml:"server"`
	Downstream  DownstreamConfig `yaml:"downstream"`
	OpenAPI     OpenAPIConfig    `yaml:"openapi"`
}

type ServerConfig struct {
	Host            string `yaml:"host"`
	HTTPPort        string `yaml:"http_port"`
	ShutdownTimeout string `yaml:"shutdown_timeout"`
}

type DownstreamConfig struct {
	AuthGRPCEndpoint    string `yaml:"auth_grpc_endpoint"`
	ProfileGRPCEndpoint string `yaml:"profile_grpc_endpoint"`
	ContentGRPCEndpoint string `yaml:"content_grpc_endpoint"`
}

type OpenAPIConfig struct {
	AuthFilePath    string `yaml:"auth_file_path"`
	ProfileFilePath string `yaml:"profile_file_path"`
	ContentFilePath string `yaml:"content_file_path"`
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
	cfg.Downstream.AuthGRPCEndpoint = getEnv("API_GATEWAY_AUTH_GRPC_ENDPOINT", getEnv("AUTH_GRPC_ENDPOINT", cfg.Downstream.AuthGRPCEndpoint))
	cfg.Downstream.ProfileGRPCEndpoint = getEnv("API_GATEWAY_PROFILE_GRPC_ENDPOINT", getEnv("PROFILE_GRPC_ENDPOINT", cfg.Downstream.ProfileGRPCEndpoint))
	cfg.Downstream.ContentGRPCEndpoint = getEnv("API_GATEWAY_CONTENT_GRPC_ENDPOINT", getEnv("CONTENT_GRPC_ENDPOINT", cfg.Downstream.ContentGRPCEndpoint))

	return cfg, nil
}

func setDefaults(cfg *Config) {
	if cfg.ServiceName == "" {
		cfg.ServiceName = "api-gateway"
	}
	if cfg.Server.Host == "" {
		cfg.Server.Host = "0.0.0.0"
	}
	if cfg.Server.HTTPPort == "" {
		cfg.Server.HTTPPort = "8080"
	}
	if cfg.Server.ShutdownTimeout == "" {
		cfg.Server.ShutdownTimeout = "10s"
	}
	if cfg.Downstream.AuthGRPCEndpoint == "" {
		cfg.Downstream.AuthGRPCEndpoint = "localhost:9091"
	}
	if cfg.Downstream.ProfileGRPCEndpoint == "" {
		cfg.Downstream.ProfileGRPCEndpoint = "localhost:9092"
	}
	if cfg.Downstream.ContentGRPCEndpoint == "" {
		cfg.Downstream.ContentGRPCEndpoint = "localhost:9093"
	}
	if cfg.OpenAPI.AuthFilePath == "" {
		cfg.OpenAPI.AuthFilePath = "grpc/gen/openapiv2/auth/v1/auth.swagger.json"
	}
	if cfg.OpenAPI.ProfileFilePath == "" {
		cfg.OpenAPI.ProfileFilePath = "grpc/gen/openapiv2/profile/v1/profile.swagger.json"
	}
	if cfg.OpenAPI.ContentFilePath == "" {
		cfg.OpenAPI.ContentFilePath = "grpc/gen/openapiv2/content/v1/content.swagger.json"
	}
}

func getEnv(key string, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	return value
}

func (cfg ServerConfig) HTTPAddress() string {
	return net.JoinHostPort(cfg.Host, cfg.HTTPPort)
}

func (cfg ServerConfig) ShutdownTimeoutDuration() (time.Duration, error) {
	return time.ParseDuration(cfg.ShutdownTimeout)
}
