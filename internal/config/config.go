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
