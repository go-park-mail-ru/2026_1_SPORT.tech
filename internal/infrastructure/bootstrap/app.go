package bootstrap

import (
	"database/sql"
	"fmt"
	"net/http"

	httpadapter "github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/adapters/http"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/infrastructure/config"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/infrastructure/persistence/postgres"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/usecase"
)

func Run(cfg config.Config, db *sql.DB) error {
	sportTypeRepository := postgres.NewSportTypeRepository(db)
	sportTypeUseCase := usecase.NewSportTypeUseCase(sportTypeRepository)

	sessionTTL, err := cfg.Auth.SessionTTLDuration()
	if err != nil {
		return fmt.Errorf("new session use case: %w", err)
	}
	sessionRepository := postgres.NewSessionRepository(db)
	sessionUseCase := usecase.NewSessionUseCase(sessionRepository, sessionTTL)

	userRepository := postgres.NewUserRepository(db)
	userUseCase := usecase.NewUserUseCase(userRepository)
	postRepository := postgres.NewPostRepository(db)
	postUseCase := usecase.NewPostUseCase(postRepository)

	httpHandler := httpadapter.NewHandler(httpadapter.Deps{
		SportTypeUseCase: sportTypeUseCase,
		SessionUseCase:   sessionUseCase,
		UserUseCase:      userUseCase,
		PostUseCase:      postUseCase,
		AuthCookieName:   cfg.Auth.CookieName,
	})

	server := &http.Server{
		Addr:    cfg.Server.Address(),
		Handler: httpHandler.Routes(),
	}

	return server.ListenAndServe()
}
