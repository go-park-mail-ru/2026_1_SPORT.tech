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
	GRPCPort        string `yaml:"grpc_port"`
	HTTPPort        string `yaml:"http_port"`
	ShutdownTimeout string `yaml:"shutdown_timeout"`
}

type DownstreamConfig struct {
	AuthGRPCEndpoint    string `yaml:"auth_grpc_endpoint"`
	ProfileGRPCEndpoint string `yaml:"profile_grpc_endpoint"`
	ContentGRPCEndpoint string `yaml:"content_grpc_endpoint"`
}

type OpenAPIConfig struct {
	GatewayFilePath string `yaml:"gateway_file_path"`
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
	if cfg.Server.GRPCPort == "" {
		cfg.Server.GRPCPort = "9090"
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
	if cfg.OpenAPI.GatewayFilePath == "" {
		cfg.OpenAPI.GatewayFilePath = "grpc/gen/openapiv2/gateway/v1/gateway.swagger.json"
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
