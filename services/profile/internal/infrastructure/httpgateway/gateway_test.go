package httpgateway_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	grpcadapter "github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/adapters/grpc"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/domain"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/infrastructure/httpgateway"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/usecase"
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

func TestNewLocalMuxExposesGeneratedGetProfileEndpoint(t *testing.T) {
	now := time.Date(2026, time.April, 18, 12, 0, 0, 0, time.UTC)
	handler := grpcadapter.NewServer(stubProfileUseCase{
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

	mux, err := httpgateway.NewLocalMux(context.Background(), handler)
	if err != nil {
		t.Fatalf("new local mux: %v", err)
	}

	server := httptest.NewServer(mux)
	defer server.Close()

	response, err := http.Get(server.URL + "/v1/profiles/7")
	if err != nil {
		t.Fatalf("get profile: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code: %d", response.StatusCode)
	}

	var payload struct {
		Profile struct {
			UserID string `json:"userId"`
		} `json:"profile"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Profile.UserID != "7" {
		t.Fatalf("unexpected user id: %s", payload.Profile.UserID)
	}
}
