package mappers

import (
	"errors"
	"testing"
	"time"

	authv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/auth/v1"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/auth/internal/domain"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/auth/internal/usecase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRegisterRequestToCommand(t *testing.T) {
	command, err := RegisterRequestToCommand(&authv1.RegisterRequest{
		Email:    "coach@example.com",
		Username: "coach",
		Password: "pass1234",
		Role:     authv1.UserRole_USER_ROLE_TRAINER,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if command.Email != "coach@example.com" ||
		command.Username != "coach" ||
		command.Password != "pass1234" ||
		command.Role != domain.RoleTrainer {
		t.Fatalf("unexpected command: %+v", command)
	}
}

func TestRegisterRequestToCommandRejectsUnknownRole(t *testing.T) {
	_, err := RegisterRequestToCommand(&authv1.RegisterRequest{Role: authv1.UserRole(99)})
	if !errors.Is(err, domain.ErrInvalidRole) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRequestMappers(t *testing.T) {
	login := LoginRequestToCommand(&authv1.LoginRequest{Email: "u@example.com", Password: "pass"})
	if login.Email != "u@example.com" || login.Password != "pass" {
		t.Fatalf("unexpected login command: %+v", login)
	}

	logout := LogoutRequestToCommand(&authv1.LogoutRequest{SessionToken: "token"})
	if logout.SessionToken != "token" {
		t.Fatalf("unexpected logout command: %+v", logout)
	}

	query := GetSessionRequestToQuery(&authv1.GetSessionRequest{SessionToken: "session"})
	if query.SessionToken != "session" {
		t.Fatalf("unexpected session query: %+v", query)
	}
}

func TestResponseMappers(t *testing.T) {
	expiresAt := time.Date(2026, time.May, 6, 12, 0, 0, 0, time.UTC)
	result := usecase.AuthResult{
		Account: domain.Account{
			ID:       1001,
			Email:    "trainer@example.com",
			Username: "trainer",
			Role:     domain.RoleTrainer,
			Status:   domain.StatusActive,
		},
		SessionToken:     "token",
		SessionExpiresAt: expiresAt,
	}

	response := NewAuthSessionResponse(result)
	if response.GetUser().GetUserId() != 1001 ||
		response.GetUser().GetRole() != authv1.UserRole_USER_ROLE_TRAINER ||
		response.GetSession().GetSessionToken() != "token" ||
		!response.GetSession().GetExpiresAt().AsTime().Equal(expiresAt) {
		t.Fatalf("unexpected auth response: %+v", response)
	}

	sessionResponse := NewGetSessionResponse(usecase.SessionResult{
		Account: result.Account,
		Session: domain.Session{ExpiresAt: expiresAt},
	})
	if sessionResponse.GetUser().GetEmail() != "trainer@example.com" ||
		!sessionResponse.GetSession().GetExpiresAt().AsTime().Equal(expiresAt) {
		t.Fatalf("unexpected session response: %+v", sessionResponse)
	}
}

func TestErrorToStatus(t *testing.T) {
	tests := []struct {
		name string
		err  error
		code codes.Code
	}{
		{name: "nil", err: nil, code: codes.OK},
		{name: "invalid argument", err: usecase.ErrInvalidEmail, code: codes.InvalidArgument},
		{name: "already exists", err: domain.ErrEmailTaken, code: codes.AlreadyExists},
		{name: "unauthenticated", err: domain.ErrInvalidCredentials, code: codes.Unauthenticated},
		{name: "not found", err: domain.ErrAccountNotFound, code: codes.NotFound},
		{name: "permission denied", err: domain.ErrAccountDisabled, code: codes.PermissionDenied},
		{name: "internal", err: errors.New("boom"), code: codes.Internal},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ErrorToStatus(test.err)
			if test.err == nil {
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
				return
			}
			if got := status.Code(err); got != test.code {
				t.Fatalf("code = %s, want %s", got, test.code)
			}
		})
	}
}
