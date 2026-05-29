//go:generate mockgen -source=$GOFILE -destination=profile_usecase_mock.go -package=mocks

package mocks

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/domain"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/usecase"
)

type ProfileUseCase interface {
	CreateProfile(ctx context.Context, command usecase.CreateProfileCommand) (domain.Profile, error)
	GetProfile(ctx context.Context, userID int64) (domain.Profile, error)
	GetProfileByUsername(ctx context.Context, username string) (domain.Profile, error)
	UpdateProfile(ctx context.Context, command usecase.UpdateProfileCommand) (domain.Profile, error)
	DeleteProfile(ctx context.Context, userID int64) error
	SearchAuthors(ctx context.Context, query usecase.SearchAuthorsQuery) ([]domain.AuthorSummary, error)
	UploadAvatar(ctx context.Context, command usecase.UploadAvatarCommand) (domain.Profile, error)
	DeleteAvatar(ctx context.Context, userID int64) error
	ListSportTypes(ctx context.Context) ([]domain.SportType, error)
	CreateMeasurement(ctx context.Context, command usecase.CreateMeasurementCommand) (domain.Measurement, error)
	ListMeasurements(ctx context.Context, query usecase.ListMeasurementsQuery) ([]domain.Measurement, error)
	DeleteMeasurement(ctx context.Context, command usecase.DeleteMeasurementCommand) error
}
