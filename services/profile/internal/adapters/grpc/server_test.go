package grpc_test

import (
	"context"
	"errors"
	"testing"
	"time"

	profilev1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/profile/v1"
	grpcadapter "github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/adapters/grpc"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/domain"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/mocks"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/usecase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestServerGetProfile(t *testing.T) {
	now := time.Date(2026, time.April, 18, 12, 0, 0, 0, time.UTC)
	profileUseCase := mocks.ProfileUseCase{
		CreateProfileFunc: func(ctx context.Context, command usecase.CreateProfileCommand) (domain.Profile, error) {
			return domain.Profile{}, errors.New("not implemented")
		},
		GetProfileFunc: func(ctx context.Context, userID int64) (domain.Profile, error) {
			return domain.Profile{
				UserID:    userID,
				Username:  "coach_john",
				FirstName: "John",
				LastName:  "Doe",
				IsTrainer: true,
				CreatedAt: now,
				UpdatedAt: now,
			}, nil
		},
		UpdateProfileFunc: func(ctx context.Context, command usecase.UpdateProfileCommand) (domain.Profile, error) {
			return domain.Profile{}, errors.New("not implemented")
		},
		SearchAuthorsFunc: func(ctx context.Context, query usecase.SearchAuthorsQuery) ([]domain.AuthorSummary, error) {
			return nil, errors.New("not implemented")
		},
		UploadAvatarFunc: func(ctx context.Context, command usecase.UploadAvatarCommand) (domain.Profile, error) {
			return domain.Profile{}, errors.New("not implemented")
		},
		DeleteAvatarFunc:   func(ctx context.Context, userID int64) error { return errors.New("not implemented") },
		ListSportTypesFunc: func(ctx context.Context) ([]domain.SportType, error) { return nil, errors.New("not implemented") },
	}
	server := grpcadapter.NewServer(grpcadapter.UseCases{
		Profiles: profileUseCase,
		Authors:  profileUseCase,
		Avatars:  profileUseCase,
		Sports:   profileUseCase,
	})

	response, err := server.GetProfile(context.Background(), &profilev1.GetProfileRequest{UserId: 7})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response.GetProfile().GetUserId() != 7 {
		t.Fatalf("unexpected user id: %d", response.GetProfile().GetUserId())
	}
}

func TestServerGetProfileMapsNotFound(t *testing.T) {
	profileUseCase := mocks.ProfileUseCase{
		CreateProfileFunc: func(ctx context.Context, command usecase.CreateProfileCommand) (domain.Profile, error) {
			return domain.Profile{}, nil
		},
		GetProfileFunc: func(ctx context.Context, userID int64) (domain.Profile, error) {
			return domain.Profile{}, domain.ErrProfileNotFound
		},
		UpdateProfileFunc: func(ctx context.Context, command usecase.UpdateProfileCommand) (domain.Profile, error) {
			return domain.Profile{}, nil
		},
		SearchAuthorsFunc: func(ctx context.Context, query usecase.SearchAuthorsQuery) ([]domain.AuthorSummary, error) {
			return nil, nil
		},
		UploadAvatarFunc: func(ctx context.Context, command usecase.UploadAvatarCommand) (domain.Profile, error) {
			return domain.Profile{}, nil
		},
		DeleteAvatarFunc:   func(ctx context.Context, userID int64) error { return nil },
		ListSportTypesFunc: func(ctx context.Context) ([]domain.SportType, error) { return nil, nil },
	}
	server := grpcadapter.NewServer(grpcadapter.UseCases{
		Profiles: profileUseCase,
		Authors:  profileUseCase,
		Avatars:  profileUseCase,
		Sports:   profileUseCase,
	})

	_, err := server.GetProfile(context.Background(), &profilev1.GetProfileRequest{UserId: 7})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("unexpected status code: %s", status.Code(err))
	}
}
