package main

import (
	"log"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
