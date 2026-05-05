package mocks

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/auth/internal/usecase"
)

type AuthUseCase struct {
	RegisterFunc   func(ctx context.Context, command usecase.RegisterCommand) (usecase.AuthResult, error)
	LoginFunc      func(ctx context.Context, command usecase.LoginCommand) (usecase.AuthResult, error)
	LogoutFunc     func(ctx context.Context, command usecase.LogoutCommand) error
	GetSessionFunc func(ctx context.Context, query usecase.GetSessionQuery) (usecase.SessionResult, error)
}

func (mock AuthUseCase) Register(ctx context.Context, command usecase.RegisterCommand) (usecase.AuthResult, error) {
	return mock.RegisterFunc(ctx, command)
}

func (mock AuthUseCase) Login(ctx context.Context, command usecase.LoginCommand) (usecase.AuthResult, error) {
	return mock.LoginFunc(ctx, command)
}

func (mock AuthUseCase) Logout(ctx context.Context, command usecase.LogoutCommand) error {
	return mock.LogoutFunc(ctx, command)
}

func (mock AuthUseCase) GetSession(ctx context.Context, query usecase.GetSessionQuery) (usecase.SessionResult, error) {
	return mock.GetSessionFunc(ctx, query)
}
