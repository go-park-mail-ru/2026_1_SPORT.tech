package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewConfigDefaultsAndEnvOverride(t *testing.T) {
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("MINIO_ACCESS_KEY", "minio-user")
	t.Setenv("MINIO_SECRET_KEY", "minio-pass")
	t.Setenv("STORAGE_PUBLIC_BASE_URL", "http://example.com/avatars")

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")
	if err := os.WriteFile(path, []byte("server:\n  port: \"8080\"\npostgres:\n  host: db\n  user: postgres\n  db_name: postgres\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := NewConfig(path)
	if err != nil {
		t.Fatalf("NewConfig: %v", err)
	}

	if cfg.Postgres.Password != "secret" {
		t.Fatalf("unexpected postgres password: %q", cfg.Postgres.Password)
	}
	if cfg.Storage.Port != "8000" {
		t.Fatalf("unexpected storage port: %q", cfg.Storage.Port)
	}
	if cfg.Storage.PublicBaseURL != "http://example.com/avatars" {
		t.Fatalf("unexpected public base url: %q", cfg.Storage.PublicBaseURL)
	}
	if cfg.Storage.AccessKey != "minio-user" || cfg.Storage.SecretKey != "minio-pass" {
		t.Fatalf("unexpected storage credentials: %+v", cfg.Storage)
	}
}

func TestConfigHelpers(t *testing.T) {
	cfg := Config{
		Server:   ServerConfig{Port: "8080"},
		Postgres: PostgresConfig{Host: "db", Port: "5432", User: "postgres", Password: "secret", Name: "app"},
		Auth:     AuthConfig{SessionTTL: "2h"},
		Storage:  StorageConfig{Host: "minio", Port: "8000"},
	}

	if cfg.Server.Address() != ":8080" {
		t.Fatalf("unexpected address: %q", cfg.Server.Address())
	}
	if !strings.Contains(cfg.Postgres.DSN(), "host=db") || !strings.Contains(cfg.Postgres.DSN(), "password=secret") {
		t.Fatalf("unexpected dsn: %q", cfg.Postgres.DSN())
	}
	duration, err := cfg.Auth.SessionTTLDuration()
	if err != nil || duration.String() != "2h0m0s" {
		t.Fatalf("unexpected duration: %v %v", duration, err)
	}
	if cfg.Storage.Endpoint() != "minio:8000" {
		t.Fatalf("unexpected endpoint: %q", cfg.Storage.Endpoint())
	}
}
