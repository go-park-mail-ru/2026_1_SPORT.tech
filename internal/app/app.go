package app

import (
	"database/sql"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/config"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/handler"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/repository"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/service"
)

func Run(config config.Config, db *sql.DB) error {
	sportTypeRepository := repository.NewSportTypeRepository(db)
	sportTypeService := service.NewSportTypeService(sportTypeRepository)

	httpHandler := handler.NewHandler(handler.Deps{
		SportTypeService: sportTypeService,
	})

	server := &http.Server{
		Addr:    config.Server.Address(),
		Handler: httpHandler.Routes(),
	}

	return server.ListenAndServe()
}
