package app

import (
	"database/sql"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/config"
)

func Run(config config.Config, db *sql.DB) error {
	server := &http.Server{
		Addr:    config.Server.Address(),
		Handler: NewRouter(),
	}

	return server.ListenAndServe()
}
