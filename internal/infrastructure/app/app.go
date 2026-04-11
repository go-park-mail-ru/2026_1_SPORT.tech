package app

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"

	httpadapter "github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/adapters/http"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/infrastructure/config"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/infrastructure/persistence/postgres"
	minioadapter "github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/infrastructure/storage/minio"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/usecase"
)

func Run(cfg config.Config, db *sql.DB, logger *slog.Logger) error {
	sportTypeRepository := postgres.NewSportTypeRepository(db, logger)
	sportTypeUseCase := usecase.NewSportTypeUseCase(sportTypeRepository)

	sessionTTL, err := cfg.Auth.SessionTTLDuration()
	if err != nil {
		return fmt.Errorf("new session use case: %w", err)
	}
	sessionRepository := postgres.NewSessionRepository(db, logger)
	sessionUseCase := usecase.NewSessionUseCase(sessionRepository, sessionTTL)

	userRepository := postgres.NewUserRepository(db, logger)
	avatarStorage, err := minioadapter.NewAvatarStorage(cfg.Storage)
	if err != nil {
		return fmt.Errorf("new avatar storage: %w", err)
	}
	userUseCase := usecase.NewUserUseCase(userRepository, avatarStorage)
	postRepository := postgres.NewPostRepository(db, logger)
	postUseCase := usecase.NewPostUseCase(postRepository)
	donationRepository := postgres.NewDonationRepository(db, logger)
	donationUseCase := usecase.NewDonationUseCase(donationRepository)

	httpHandler := httpadapter.NewHandler(httpadapter.Deps{
		Logger:           logger,
		SportTypeUseCase: sportTypeUseCase,
		SessionUseCase:   sessionUseCase,
		UserUseCase:      userUseCase,
		PostUseCase:      postUseCase,
		DonationUseCase:  donationUseCase,
		AuthCookieName:   cfg.Auth.CookieName,
	})

	server := &http.Server{
		Addr:    cfg.Server.Address(),
		Handler: httpHandler.Routes(),
	}

	return server.ListenAndServe()
}
