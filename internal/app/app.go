package app

import (
	"net/http"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/config"
)

func Run(config config.Config) error {
	server := &http.Server{
		Addr:    config.Server.Address(),
		Handler: NewRouter(),
	}

	return server.ListenAndServe()
}
