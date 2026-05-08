package mocks

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/domain"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/usecase"
)

type ProfileUseCase struct {
	CreateProfileFunc  func(ctx context.Context, command usecase.CreateProfileCommand) (domain.Profile, error)
	GetProfileFunc     func(ctx context.Context, userID int64) (domain.Profile, error)
	UpdateProfileFunc  func(ctx context.Context, command usecase.UpdateProfileCommand) (domain.Profile, error)
	SearchAuthorsFunc  func(ctx context.Context, query usecase.SearchAuthorsQuery) ([]domain.AuthorSummary, error)
	UploadAvatarFunc   func(ctx context.Context, command usecase.UploadAvatarCommand) (domain.Profile, error)
	DeleteAvatarFunc   func(ctx context.Context, userID int64) error
	ListSportTypesFunc func(ctx context.Context) ([]domain.SportType, error)
}

func (mock ProfileUseCase) CreateProfile(ctx context.Context, command usecase.CreateProfileCommand) (domain.Profile, error) {
	return mock.CreateProfileFunc(ctx, command)
}

func (mock ProfileUseCase) GetProfile(ctx context.Context, userID int64) (domain.Profile, error) {
	return mock.GetProfileFunc(ctx, userID)
}

func (mock ProfileUseCase) UpdateProfile(ctx context.Context, command usecase.UpdateProfileCommand) (domain.Profile, error) {
	return mock.UpdateProfileFunc(ctx, command)
}

func (mock ProfileUseCase) SearchAuthors(ctx context.Context, query usecase.SearchAuthorsQuery) ([]domain.AuthorSummary, error) {
	return mock.SearchAuthorsFunc(ctx, query)
}

func (mock ProfileUseCase) UploadAvatar(ctx context.Context, command usecase.UploadAvatarCommand) (domain.Profile, error) {
	return mock.UploadAvatarFunc(ctx, command)
}

func (mock ProfileUseCase) DeleteAvatar(ctx context.Context, userID int64) error {
	return mock.DeleteAvatarFunc(ctx, userID)
}

func (mock ProfileUseCase) ListSportTypes(ctx context.Context) ([]domain.SportType, error) {
	return mock.ListSportTypesFunc(ctx)
}
