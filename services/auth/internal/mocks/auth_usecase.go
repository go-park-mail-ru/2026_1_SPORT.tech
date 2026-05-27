//go:generate mockgen -source=$GOFILE -destination=auth_usecase_mock.go -package=mocks

package mocks

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/auth/internal/usecase"
)

type AuthUseCase interface {
	Register(ctx context.Context, command usecase.RegisterCommand) (usecase.AuthResult, error)
	Login(ctx context.Context, command usecase.LoginCommand) (usecase.AuthResult, error)
	Logout(ctx context.Context, command usecase.LogoutCommand) error
	GetSession(ctx context.Context, query usecase.GetSessionQuery) (usecase.SessionResult, error)
}
