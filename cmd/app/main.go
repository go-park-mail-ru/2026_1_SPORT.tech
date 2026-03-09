package main

import (
	"log"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/app"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/config"
)

const configPath = "config.yml"

func main() {
	cfg, err := config.NewConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	if err := app.Run(cfg); err != nil {
		log.Fatal(err)
	}
}
