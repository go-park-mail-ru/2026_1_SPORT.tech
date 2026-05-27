package httpgateway_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	grpcadapter "github.com/go-park-mail-ru/2026_1_SPORT.tech/services/auth/internal/adapters/grpc"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/auth/internal/domain"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/auth/internal/infrastructure/httpgateway"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/auth/internal/mocks"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/auth/internal/usecase"
	"go.uber.org/mock/gomock"
)

func TestNewLocalMuxExposesGeneratedLoginEndpoint(t *testing.T) {
	now := time.Date(2026, time.April, 18, 12, 0, 0, 0, time.UTC)
	var capturedCommand usecase.LoginCommand

	ctrl := gomock.NewController(t)
	authUseCase := mocks.NewMockAuthUseCase(ctrl)
	authUseCase.EXPECT().
		Login(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, command usecase.LoginCommand) (usecase.AuthResult, error) {
			capturedCommand = command
			return usecase.AuthResult{
				Account: domain.Account{
					ID:       7,
					Email:    command.Email,
					Username: "john_doe",
					Role:     domain.RoleClient,
					Status:   domain.StatusActive,
				},
				SessionToken:     "session-token",
				SessionExpiresAt: now.Add(24 * time.Hour),
			}, nil
		})

	handler := grpcadapter.NewServer(grpcadapter.UseCases{
		Registration: authUseCase,
		Login:        authUseCase,
		Session:      authUseCase,
	})

	mux, err := httpgateway.NewLocalMux(context.Background(), handler)
	if err != nil {
		t.Fatalf("new local mux: %v", err)
	}

	server := httptest.NewServer(mux)
	defer server.Close()

	requestBody := map[string]string{
		"email":    "john@example.com",
		"password": "supersecret123",
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	response, err := http.Post(server.URL+"/v1/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("post login: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code: %d", response.StatusCode)
	}

	var payload struct {
		User struct {
			UserID string `json:"userId"`
			Email  string `json:"email"`
		} `json:"user"`
		Session struct {
			SessionToken string `json:"sessionToken"`
		} `json:"session"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload.User.UserID != "7" {
		t.Fatalf("unexpected user id: %s", payload.User.UserID)
	}
	if payload.Session.SessionToken != "session-token" {
		t.Fatalf("unexpected session token: %s", payload.Session.SessionToken)
	}
	if capturedCommand.Email != "john@example.com" {
		t.Fatalf("unexpected captured email: %s", capturedCommand.Email)
	}
}
