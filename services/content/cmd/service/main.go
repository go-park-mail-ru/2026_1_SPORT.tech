package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/infrastructure/bootstrap"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/infrastructure/config"
)

const defaultConfigPath = "services/content/configs/service.yml"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	configPath := os.Getenv("CONTENT_CONFIG_PATH")
	if configPath == "" {
		configPath = defaultConfigPath
	}

	cfg, err := config.NewConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	app, err := bootstrap.New(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}

	if err := app.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
