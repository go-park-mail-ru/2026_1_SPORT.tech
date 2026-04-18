package grpc_test

import (
	"context"
	"errors"
	"testing"
	"time"

	authv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/auth/v1"
	grpcadapter "github.com/go-park-mail-ru/2026_1_SPORT.tech/services/auth/internal/adapters/grpc"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/auth/internal/domain"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/auth/internal/usecase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type stubAuthUseCase struct {
	registerFunc   func(ctx context.Context, command usecase.RegisterCommand) (usecase.AuthResult, error)
	loginFunc      func(ctx context.Context, command usecase.LoginCommand) (usecase.AuthResult, error)
	logoutFunc     func(ctx context.Context, command usecase.LogoutCommand) error
	getSessionFunc func(ctx context.Context, query usecase.GetSessionQuery) (usecase.SessionResult, error)
}

func (stub stubAuthUseCase) Register(ctx context.Context, command usecase.RegisterCommand) (usecase.AuthResult, error) {
	return stub.registerFunc(ctx, command)
}

func (stub stubAuthUseCase) Login(ctx context.Context, command usecase.LoginCommand) (usecase.AuthResult, error) {
	return stub.loginFunc(ctx, command)
}

func (stub stubAuthUseCase) Logout(ctx context.Context, command usecase.LogoutCommand) error {
	return stub.logoutFunc(ctx, command)
}

func (stub stubAuthUseCase) GetSession(ctx context.Context, query usecase.GetSessionQuery) (usecase.SessionResult, error) {
	return stub.getSessionFunc(ctx, query)
}

func TestServerRegister(t *testing.T) {
	now := time.Date(2026, time.April, 18, 12, 0, 0, 0, time.UTC)

	server := grpcadapter.NewServer(stubAuthUseCase{
		registerFunc: func(ctx context.Context, command usecase.RegisterCommand) (usecase.AuthResult, error) {
			if command.Email != "john@example.com" {
				t.Fatalf("unexpected email: %s", command.Email)
			}
			if command.Role != domain.RoleTrainer {
				t.Fatalf("unexpected role: %s", command.Role)
			}

			return usecase.AuthResult{
				Account: domain.Account{
					ID:       11,
					Email:    command.Email,
					Username: command.Username,
					Role:     command.Role,
					Status:   domain.StatusActive,
				},
				SessionToken:     "session-token",
				SessionExpiresAt: now.Add(24 * time.Hour),
			}, nil
		},
		loginFunc: func(ctx context.Context, command usecase.LoginCommand) (usecase.AuthResult, error) {
			return usecase.AuthResult{}, errors.New("not implemented")
		},
		logoutFunc: func(ctx context.Context, command usecase.LogoutCommand) error { return errors.New("not implemented") },
		getSessionFunc: func(ctx context.Context, query usecase.GetSessionQuery) (usecase.SessionResult, error) {
			return usecase.SessionResult{}, errors.New("not implemented")
		},
	})

	response, err := server.Register(context.Background(), &authv1.RegisterRequest{
		Email:    "john@example.com",
		Username: "coach_john",
		Password: "supersecret123",
		Role:     authv1.UserRole_USER_ROLE_TRAINER,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response.GetUser().GetUserId() != 11 {
		t.Fatalf("unexpected user id: %d", response.GetUser().GetUserId())
	}
	if response.GetSession().GetSessionToken() != "session-token" {
		t.Fatalf("unexpected session token: %s", response.GetSession().GetSessionToken())
	}
}

func TestServerRegisterInvalidRole(t *testing.T) {
	server := grpcadapter.NewServer(stubAuthUseCase{
		registerFunc: func(ctx context.Context, command usecase.RegisterCommand) (usecase.AuthResult, error) {
			return usecase.AuthResult{}, nil
		},
		loginFunc: func(ctx context.Context, command usecase.LoginCommand) (usecase.AuthResult, error) {
			return usecase.AuthResult{}, nil
		},
		logoutFunc: func(ctx context.Context, command usecase.LogoutCommand) error { return nil },
		getSessionFunc: func(ctx context.Context, query usecase.GetSessionQuery) (usecase.SessionResult, error) {
			return usecase.SessionResult{}, nil
		},
	})

	_, err := server.Register(context.Background(), &authv1.RegisterRequest{
		Email:    "john@example.com",
		Username: "coach_john",
		Password: "supersecret123",
		Role:     authv1.UserRole(99),
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("unexpected status code: %s", status.Code(err))
	}
}
