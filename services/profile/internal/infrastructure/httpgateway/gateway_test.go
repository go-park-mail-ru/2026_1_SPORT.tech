package httpgateway_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	grpcadapter "github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/adapters/grpc"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/domain"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/infrastructure/httpgateway"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/mocks"
	"go.uber.org/mock/gomock"
)

func TestNewLocalMuxExposesGeneratedGetProfileEndpoint(t *testing.T) {
	now := time.Date(2026, time.April, 18, 12, 0, 0, 0, time.UTC)
	ctrl := gomock.NewController(t)
	profileUseCase := mocks.NewMockProfileUseCase(ctrl)
	profileUseCase.EXPECT().
		GetProfile(gomock.Any(), int64(7)).
		Return(domain.Profile{
			UserID:    7,
			Username:  "coach_john",
			FirstName: "John",
			LastName:  "Doe",
			IsTrainer: true,
			CreatedAt: now,
			UpdatedAt: now,
		}, nil)

	handler := grpcadapter.NewServer(grpcadapter.UseCases{
		Profiles: profileUseCase,
		Authors:  profileUseCase,
		Avatars:  profileUseCase,
		Sports:   profileUseCase,
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
