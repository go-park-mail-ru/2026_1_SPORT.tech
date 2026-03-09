package app

import (
	"database/sql"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/config"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/handler"
)

func Run(config config.Config, db *sql.DB) error {
	_ = db

	httpHandler := handler.NewHandler(handler.Deps{})

	server := &http.Server{
		Addr:    config.Server.Address(),
		Handler: httpHandler.Routes(),
	}

	return server.ListenAndServe()
}
