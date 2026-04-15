package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Postgres PostgresConfig `yaml:"postgres"`
	Auth     AuthConfig     `yaml:"auth"`
	Storage  StorageConfig  `yaml:"storage"`
}

type ServerConfig struct {
	Port string `yaml:"port"`
}

type PostgresConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string
	Name     string `yaml:"db_name"`
}

type AuthConfig struct {
	CookieName string `yaml:"cookie_name"`
	SessionTTL string `yaml:"session_ttl"`
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

func NewConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}

	config.Postgres.Password = getEnv("DB_PASSWORD", "postgres")
	if config.Auth.CookieName == "" {
		config.Auth.CookieName = "sid"
	}
	if config.Auth.SessionTTL == "" {
		config.Auth.SessionTTL = "720h"
	}
	if config.Storage.Host == "" {
		config.Storage.Host = "minio"
	}
	if config.Storage.Port == "" {
		config.Storage.Port = "8000"
	}
	if config.Storage.Bucket == "" {
		config.Storage.Bucket = "avatars"
	}
	if config.Storage.PublicBaseURL == "" {
		config.Storage.PublicBaseURL = "http://localhost:8000/avatars"
	}
	config.Storage.PublicBaseURL = getEnv("STORAGE_PUBLIC_BASE_URL", config.Storage.PublicBaseURL)
	config.Storage.AccessKey = getEnv("MINIO_ACCESS_KEY", "minioadmin")
	config.Storage.SecretKey = getEnv("MINIO_SECRET_KEY", "minioadmin")

	return config, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

func (config PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.Name,
	)
}

func (config ServerConfig) Address() string {
	return ":" + config.Port
}

func (config AuthConfig) SessionTTLDuration() (time.Duration, error) {
	return time.ParseDuration(config.SessionTTL)
}

func (config StorageConfig) Endpoint() string {
	return fmt.Sprintf("%s:%s", config.Host, config.Port)
}
