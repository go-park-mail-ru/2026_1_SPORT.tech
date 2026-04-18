package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/auth/internal/infrastructure/config"
	_ "github.com/lib/pq"
)

const pingTimeout = 5 * time.Second

func NewPostgres(cfg config.PostgresConfig) (*sql.DB, error) {
	database, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	database.SetMaxOpenConns(cfg.MaxOpenConns)
	database.SetMaxIdleConns(cfg.MaxIdleConns)

	connMaxLifetime, err := cfg.ConnMaxLifetimeDuration()
	if err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("parse postgres conn max lifetime: %w", err)
	}
	database.SetConnMaxLifetime(connMaxLifetime)

	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()

	if err := database.PingContext(ctx); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return database, nil
}
