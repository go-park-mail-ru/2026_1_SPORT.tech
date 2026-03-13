package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/infrastructure/app"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/infrastructure/config"
)

const (
	configPath  = "config.yml"
	pingTimeout = 5 * time.Second
)

func main() {
	cfg, err := config.NewConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	db, err := initDB(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := app.Run(cfg, db); err != nil {
		log.Fatal(err)
	}
}

func initDB(cfg config.Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.Postgres.DSN())
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}
