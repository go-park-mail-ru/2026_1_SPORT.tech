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
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/mocks"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/usecase"
)

func TestNewLocalMuxExposesGeneratedGetProfileEndpoint(t *testing.T) {
	now := time.Date(2026, time.April, 18, 12, 0, 0, 0, time.UTC)
	handler := grpcadapter.NewServer(mocks.ProfileUseCase{
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
