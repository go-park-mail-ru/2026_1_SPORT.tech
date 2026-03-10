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
	sportTypeRepository := repository.NewSportTypeRepository(db)
	sportTypeService := service.NewSportTypeService(sportTypeRepository)

	sessionRepository := repository.NewSessionRepository(db)
	sessionService, err := service.NewSessionService(sessionRepository, config.Auth)
	if err != nil {
		return fmt.Errorf("new session service: %w", err)
	}

	userRepository := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepository)
	postRepository := repository.NewPostRepository(db)
	postService := service.NewPostService(postRepository)

	httpHandler := handler.NewHandler(handler.Deps{
		SportTypeService: sportTypeService,
		SessionService:   sessionService,
		UserService:      userService,
		PostService:      postService,
		AuthCookieName:   config.Auth.CookieName,
	})

	server := &http.Server{
		Addr:    config.Server.Address(),
		Handler: httpHandler.Routes(),
	}

	return server.ListenAndServe()
}
