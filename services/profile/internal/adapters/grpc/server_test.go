package grpc_test

import (
	"context"
	"errors"
	"testing"
	"time"

	profilev1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/profile/v1"
	grpcadapter "github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/adapters/grpc"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/domain"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/usecase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type stubProfileUseCase struct {
	createProfileFunc  func(ctx context.Context, command usecase.CreateProfileCommand) (domain.Profile, error)
	getProfileFunc     func(ctx context.Context, userID int64) (domain.Profile, error)
	updateProfileFunc  func(ctx context.Context, command usecase.UpdateProfileCommand) (domain.Profile, error)
	searchAuthorsFunc  func(ctx context.Context, query usecase.SearchAuthorsQuery) ([]domain.AuthorSummary, error)
	uploadAvatarFunc   func(ctx context.Context, command usecase.UploadAvatarCommand) (domain.Profile, error)
	deleteAvatarFunc   func(ctx context.Context, userID int64) error
	listSportTypesFunc func(ctx context.Context) ([]domain.SportType, error)
}

func (stub stubProfileUseCase) CreateProfile(ctx context.Context, command usecase.CreateProfileCommand) (domain.Profile, error) {
	return stub.createProfileFunc(ctx, command)
}

func (stub stubProfileUseCase) GetProfile(ctx context.Context, userID int64) (domain.Profile, error) {
	return stub.getProfileFunc(ctx, userID)
}

func (stub stubProfileUseCase) UpdateProfile(ctx context.Context, command usecase.UpdateProfileCommand) (domain.Profile, error) {
	return stub.updateProfileFunc(ctx, command)
}

func (stub stubProfileUseCase) SearchAuthors(ctx context.Context, query usecase.SearchAuthorsQuery) ([]domain.AuthorSummary, error) {
	return stub.searchAuthorsFunc(ctx, query)
}

func (stub stubProfileUseCase) UploadAvatar(ctx context.Context, command usecase.UploadAvatarCommand) (domain.Profile, error) {
	return stub.uploadAvatarFunc(ctx, command)
}

func (stub stubProfileUseCase) DeleteAvatar(ctx context.Context, userID int64) error {
	return stub.deleteAvatarFunc(ctx, userID)
}

func (stub stubProfileUseCase) ListSportTypes(ctx context.Context) ([]domain.SportType, error) {
	return stub.listSportTypesFunc(ctx)
}

func TestServerGetProfile(t *testing.T) {
	now := time.Date(2026, time.April, 18, 12, 0, 0, 0, time.UTC)
	server := grpcadapter.NewServer(stubProfileUseCase{
		createProfileFunc: func(ctx context.Context, command usecase.CreateProfileCommand) (domain.Profile, error) {
			return domain.Profile{}, errors.New("not implemented")
		},
		getProfileFunc: func(ctx context.Context, userID int64) (domain.Profile, error) {
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
		updateProfileFunc: func(ctx context.Context, command usecase.UpdateProfileCommand) (domain.Profile, error) {
			return domain.Profile{}, errors.New("not implemented")
		},
		searchAuthorsFunc: func(ctx context.Context, query usecase.SearchAuthorsQuery) ([]domain.AuthorSummary, error) {
			return nil, errors.New("not implemented")
		},
		uploadAvatarFunc: func(ctx context.Context, command usecase.UploadAvatarCommand) (domain.Profile, error) {
			return domain.Profile{}, errors.New("not implemented")
		},
		deleteAvatarFunc:   func(ctx context.Context, userID int64) error { return errors.New("not implemented") },
		listSportTypesFunc: func(ctx context.Context) ([]domain.SportType, error) { return nil, errors.New("not implemented") },
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
	server := grpcadapter.NewServer(stubProfileUseCase{
		createProfileFunc: func(ctx context.Context, command usecase.CreateProfileCommand) (domain.Profile, error) {
			return domain.Profile{}, nil
		},
		getProfileFunc: func(ctx context.Context, userID int64) (domain.Profile, error) {
			return domain.Profile{}, domain.ErrProfileNotFound
		},
		updateProfileFunc: func(ctx context.Context, command usecase.UpdateProfileCommand) (domain.Profile, error) {
			return domain.Profile{}, nil
		},
		searchAuthorsFunc: func(ctx context.Context, query usecase.SearchAuthorsQuery) ([]domain.AuthorSummary, error) {
			return nil, nil
		},
		uploadAvatarFunc: func(ctx context.Context, command usecase.UploadAvatarCommand) (domain.Profile, error) {
			return domain.Profile{}, nil
		},
		deleteAvatarFunc:   func(ctx context.Context, userID int64) error { return nil },
		listSportTypesFunc: func(ctx context.Context) ([]domain.SportType, error) { return nil, nil },
	})

	_, err := server.GetProfile(context.Background(), &profilev1.GetProfileRequest{UserId: 7})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("unexpected status code: %s", status.Code(err))
	}
}
