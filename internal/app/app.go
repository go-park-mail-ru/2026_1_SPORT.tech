package app

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/config"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/handler"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/repository"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/service"
)

func Run(config config.Config, db *sql.DB) error {
	sessionRepository := repository.NewSessionRepository(db)
	sessionService, err := service.NewSessionService(sessionRepository, config.Auth)
	if err != nil {
		return fmt.Errorf("new session service: %w", err)
	}

	httpHandler := handler.NewHandler(handler.Deps{
		SessionService: sessionService,
		AuthCookieName: config.Auth.CookieName,
	})

	server := &http.Server{
		Addr:    config.Server.Address(),
		Handler: httpHandler.Routes(),
	}

	return server.ListenAndServe()
}
